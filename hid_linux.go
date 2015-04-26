package hid

// #cgo pkg-config: libusb-1.0
// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import (
	"errors"
	"fmt"
	"reflect"
	"syscall"
	"unsafe"
)

type linuxDevice struct {
	handle int
}

func Init() {
	C.libusb_init(nil)
}

func newDeviceInfo(dev *C.libusb_device) (*DeviceInfo, error) {
	var desc C.struct_libusb_device_descriptor
	if errno := C.libusb_get_device_descriptor(dev, &desc); errno < 0 {
		return nil, usbError(errno)
	}

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
	handle, err := syscall.Open(di.Path, syscall.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	if handle > 0 {
		return &linuxDevice{handle}, nil
	} else {
		return nil, errors.New("unable to open file")
	}
}

func (dev *linuxDevice) Close() {
	syscall.Close(dev.handle)
}

func (dev *linuxDevice) WriteFeature(data []byte) error {
	return errors.New("not yet implemented")
}

func (dev *linuxDevice) Write(data []byte) error {
	n, err := syscall.Write(dev.handle, data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("written bytes missmatch!")
	}
	return err
}
