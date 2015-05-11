package hid

// #cgo pkg-config: libusb-1.0
// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf16"
	"unsafe"
)

type linuxDevice struct {
	handle *C.libusb_device_handle
	info   *DeviceInfo
}

func init() {
	C.libusb_init(nil)
}

func Devices() <-chan *DeviceInfo {
	result := make(chan *DeviceInfo)
	go func() {
		var devices **C.struct_libusb_device
		cnt := C.libusb_get_device_list(nil, &devices)
		if cnt < 0 {
			close(result)
			return
		}
		defer C.libusb_free_device_list(devices, 1)

		var device_list []*C.libusb_device
		*(*reflect.SliceHeader)(unsafe.Pointer(&device_list)) = reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(devices)),
			Len:  int(cnt),
			Cap:  int(cnt),
		}

		for _, dev := range device_list {
			di, err := newDeviceInfo(dev)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				continue
			}
			result <- di
		}

		close(result)
	}()
	return result
}

func ByPath(path string) (*DeviceInfo, error) {
	for d := range Devices() {
		if d.Path == path {
			return d, nil
		}
	}
	return nil, errors.New("Device not found")
}

func (di *DeviceInfo) Open() (Device, error) {
	// todo: use another mechanism, cause vid/pid isn't uniqqu for multiple devices
	// maybe libusb_get_port_numbers can be sused as path
	var h *C.libusb_device_handle
	h = C.libusb_open_device_with_vid_pid(nil, C.uint16_t(di.VendorId), C.uint16_t(di.ProductId))
	if h == nil {
		return nil, errors.New(fmt.Sprintf("Couldn't open USB device vendor=%4x product=%4x", di.VendorId, di.ProductId))
	}
	dev := &linuxDevice{
		info:   di,
		handle: h,
	}
	return dev, nil
}

func (dev *linuxDevice) Close() {
	if dev.handle != nil {
		C.libusb_close(dev.handle)
		dev.handle = nil
		dev.info = nil
	}
}

func (dev *linuxDevice) writeReport(hid_report_type int, data []byte) error {
	if dev.handle == nil {
		return errors.New("No USB device opend before.")
	}
	if len(data) > 0xffff {
		return errors.New("data longer than 65535 bytes, means overflow, isn't supported")
	}
	if len(data) == 0 {
		return nil
	}

	const reportId = 1
	const index = 0
	const timeout = 1000

	written := C.libusb_control_transfer(dev.handle,
		C.uint8_t(ENDPOINT_OUT|RECIPIENT_DEVICE|DT_REPORT|hid_report_type),
		C.uint8_t(HID_SET_REPORT),
		C.uint16_t(reportId),
		C.uint16_t(index),
		(*C.uchar)(&data[0]),
		C.uint16_t(len(data)),
		C.uint(timeout))

	if int(written) == len(data) {
		return nil
	}
	return usbError(written)
}

func (dev *linuxDevice) WriteFeature(data []byte) error {
	return dev.writeReport(HID_REPORT_TYPE_FEATURE, data)
}

func (dev *linuxDevice) Write(data []byte) error {
	return dev.writeReport(HID_REPORT_TYPE_OUTPUT, data)
}

func newDeviceInfo(dev *C.libusb_device) (*DeviceInfo, error) {
	var desc C.struct_libusb_device_descriptor
	if err := C.libusb_get_device_descriptor(dev, &desc); err < 0 {
		return nil, usbError(err)
	}
	manufacturer, product, _ := resolveDescriptors(dev, desc.iManufacturer, desc.iProduct)
	return &DeviceInfo{
		Path:          getPath(dev, desc.idVendor, desc.idProduct),
		VendorId:      uint16(desc.idVendor),
		ProductId:     uint16(desc.idProduct),
		VersionNumber: uint16(desc.bcdDevice),
		Manufacturer:  manufacturer,
		Product:       product,
	}, nil
}

func resolveDescriptors(dev *C.libusb_device, iManufacturer C.uint8_t, iProduct C.uint8_t) (manufacturer string, product string, e error) {
	var handle *C.libusb_device_handle
	err := C.libusb_open(dev, &handle)
	if handle != nil {
		defer C.libusb_close(handle)
		if err != 0 {
			return "", "", usbError(err)
		}
		manufacturerStr, _ := getStringDescriptor(handle, iManufacturer)
		productStr, _ := getStringDescriptor(handle, iProduct)
		return manufacturerStr, productStr, nil
	}
	return "", "", errors.New("Couldn't resolve description string")

}

func getStringDescriptor(dev *C.libusb_device_handle, id C.uint8_t) (string, error) {
	var buf [128]C.char
	const langId = 0
	err := C.libusb_get_string_descriptor(dev, id, C.uint16_t(langId), (*C.uchar)(unsafe.Pointer(&buf[0])), C.int(len(buf)))
	if err < 0 {
		return "", usbError(err)
	}
	if err < 2 {
		return "", errors.New("not enough data for USB string descriptor")
	}
	l := C.int(buf[0])
	if l > err {
		return "", errors.New("USB string descriptor is too short")
	}
	b := buf[2:l]
	uni16 := make([]uint16, len(b)/2)
	for i := range uni16 {
		uni16[i] = uint16(b[i*2]) | uint16(b[i*2+1])<<8
	}
	return string(utf16.Decode(uni16)), nil
}

func getPath(dev *C.libusb_device, vendorId C.uint16_t, productId C.uint16_t) string {
	numbers, _ := getPortNumbers(dev)
	path := fmt.Sprintf("%.4x:%.4x:%s", vendorId, productId, numbers)
	return path
}

func getPortNumbers(dev *C.libusb_device) (string, error) {
	const maxlen = 7 // As per the USB 3.0 specs, the current maximum limit for the depth is 7
	var numarr [maxlen]C.uint8_t
	len := C.libusb_get_port_numbers(dev, &numarr[0], maxlen)
	if len < 0 || len > maxlen {
		return "", usbError(len)
	}
	var numstr []string = make([]string, len)
	for i := 0; i < int(len); i++ {
		numstr[i] = fmt.Sprintf("%.2x", numarr[i])
	}
	return strings.Join(numstr, "."), nil
}
