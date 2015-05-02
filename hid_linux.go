package hid

// #cgo pkg-config: libusb-1.0
// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

type linuxDevice struct {
	handle *C.libusb_device_handle
	info   *DeviceInfo
}

func Init() {
	C.libusb_init(nil)
}

func newDeviceInfo(dev *C.libusb_device) (*DeviceInfo, error) {
	var desc C.struct_libusb_device_descriptor
	if errno := C.libusb_get_device_descriptor(dev, &desc); errno < 0 {
		return nil, usbError(errno)
	}
	// todo: check if libusb_get_string_descriptor can be used to get manufacturer
	return &DeviceInfo{
		Path:          "",
		VendorId:      uint16(desc.idVendor),
		ProductId:     uint16(desc.idProduct),
		VersionNumber: uint16(desc.bcdDevice),
		Manufacturer:  "",
		Product:       "",
	}, nil
}

func Devices() <-chan *DeviceInfo {
	result := make(chan *DeviceInfo)
	go func() {
		var devices **C.libusb_device
		cnt := C.libusb_get_device_list(nil, &devices)
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
	dev := &linuxDevice{
		info:   di,
		handle: C.libusb_open_device_with_vid_pid(nil, C.uint16_t(di.VendorId), C.uint16_t(di.ProductId)),
	}
	return dev, nil
}

func (dev *linuxDevice) Close() {
	C.libusb_exit(nil)
}

func (dev *linuxDevice) WriteFeature(data []byte) error {
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

	C.libusb_control_transfer(dev.handle,
		C.uint8_t(ENDPOINT_OUT|REQUEST_TYPE_CLASS|RECIPIENT_DEVICE),
		C.uint8_t(HID_SET_REPORT),
		C.uint16_t(reportId),
		C.uint16_t(index),
		(*C.uchar)(&data[0]),
		C.uint16_t(len(data)),
		C.uint(timeout))

	return nil
}

func (dev *linuxDevice) Write(data []byte) error {
	return nil
}
