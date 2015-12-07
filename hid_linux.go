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
		for _, dev := range asSlice(devices, cnt) {
			di, err := newDeviceInfo(dev)
			if err != nil {
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

	var devices **C.struct_libusb_device
	cnt := C.libusb_get_device_list(nil, &devices)
	if cnt < 0 {
		return nil, errors.New(fmt.Sprintf("Couldn't open USB device with path=%s, because couldn't enumerate USB devices.", di.Path))
	}
	defer C.libusb_free_device_list(devices, 1)

	for _, dev := range asSlice(devices, cnt) {
		candidate, err := newDeviceInfo(dev)
		if err != nil {
			continue
		}
		if di.Path == candidate.Path {
			var handle *C.libusb_device_handle
			err := C.libusb_open(dev, &handle)
			if err != 0 {
				return nil, usbError(err)
			}
			dev := &linuxDevice{
				info:   candidate,
				handle: handle,
			}
			return dev, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Couldn't open USB device vendor=%4x product=%4x", di.VendorId, di.ProductId))
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

func (dev *linuxDevice) WriteInterrupt(endpoint byte, data []byte) (int, error) {
	if dev.handle == nil {
		return 0, errors.New("No USB device opend before.")
	}
	if len(data) > 0xffff {
		return 0, errors.New("data longer than 65535 bytes, means overflow, isn't supported")
	}
	if len(data) == 0 {
		return 0, nil
	}
	const timeout = 10000
	var transferred C.int
	rval := C.libusb_interrupt_transfer(dev.handle,
		C.uchar(endpoint),
		(*C.uchar)(&data[0]),
		C.int(len(data)),
		&transferred,
		timeout)
	if rval != 0 {
		return 0, usbError(rval)
	}
	return int(transferred), nil
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
	manufacturer, product, err := resolveDescriptors(dev, desc.iManufacturer, desc.iProduct)
	if err == usbError(C.LIBUSB_ERROR_ACCESS) {
		manufacturer = "access not allowed"
		product = "access not allowed"
	} else if err != nil {
		return nil, err
	}
	path, err := getPath(dev, desc.idVendor, desc.idProduct)
	if err != nil {
		return nil, err
	}
	return &DeviceInfo{
		Path:          path,
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
	if err != 0 {
		return "", "", usbError(err)
	}
	if handle != nil {
		defer C.libusb_close(handle)
		manufacturerStr, err := getStringDescriptor(handle, iManufacturer)
		if err != nil {
			return "", "", err
		}
		productStr, err := getStringDescriptor(handle, iProduct)
		if err != nil {
			return "", "", err
		}
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

func getPath(dev *C.libusb_device, vendorId C.uint16_t, productId C.uint16_t) (string, error) {
	numbers, err := getPortNumbers(dev)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%.4x:%.4x:%s", vendorId, productId, numbers)
	return path, nil
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

func asSlice(devices **C.struct_libusb_device, cnt C.ssize_t) []*C.libusb_device {
	var device_list []*C.libusb_device
	*(*reflect.SliceHeader)(unsafe.Pointer(&device_list)) = reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(devices)),
		Len:  int(cnt),
		Cap:  int(cnt),
	}
	return device_list
}
