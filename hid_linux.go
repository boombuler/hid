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
	handle C.libusb_device_handle
	info   DeviceInfo
}

func Init() {
	C.libusb_init(nil)

	// TODO : C.libusb_exit()
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
		info:   *di,
		handle: *C.libusb_open_device_with_vid_pid(nil, C.uint16_t(di.VendorId), C.uint16_t(di.ProductId)),
	}
	return dev, nil
}

func (dev *linuxDevice) Close() {

}

func (dev *linuxDevice) ControlTransfer(reqType, req byte, value, index uint16, data []byte, timeout int) error {
	var ptr *byte
	if len(data) > 0 {
		ptr = &data[0]
	}
	if len(data) > 0xffff {
		return errors.New("data longer than 65535 bytes, means overflow")
	}
	err := usbError(C.libusb_control_transfer(&dev.handle,
		C.uint8_t(reqType), C.uint8_t(req), C.uint16_t(value), C.uint16_t(index),
		(*C.uchar)(ptr), C.uint16_t(len(data)), C.uint(timeout)))
	return errors.New(err.Error())
}

func (dev *linuxDevice) WriteFeature(data []byte) error {
	dev.ControlTransfer(
		ENDPOINT_OUT|REQUEST_TYPE_STANDARD|RECIPIENT_INTERFACE,
		REQUEST_SET_FEATURE,
		DT_REPORT,
		0,
		data,
		1000)
	//req_type := C.LIBUSB_ENDPOINT_IN | C.LIBUSB_REQUEST_TYPE_STANDARD | C.LIBUSB_RECIPIENT_INTERFACE
	//C.libusb_control_transfer(&dev.handle,
	//C.uint8_t(req_type),
	//C.uint8_t(C.LIBUSB_REQUEST_GET_DESCRIPTOR),
	//C.uint16_t(C.LIBUSB_DT_REPORT<<8),
	//C.uint16_t(0),
	//(*C.uchar)(&data[0]),
	// C.uint16_t(len(data)),
	//C.uint(1000),
	//)

	return errors.New("not yet implemented")
}

func (dev *linuxDevice) Write(data []byte) error {
	return nil
}
