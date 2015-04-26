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

type usbError C.int

func (e usbError) Error() string {
	return fmt.Sprintf("libusb: %s [code %d]", usbErrorString[e], int(e))
}

const (
	SUCCESS             usbError = C.LIBUSB_SUCCESS
	ERROR_IO            usbError = C.LIBUSB_ERROR_IO
	ERROR_INVALID_PARAM usbError = C.LIBUSB_ERROR_INVALID_PARAM
	ERROR_ACCESS        usbError = C.LIBUSB_ERROR_ACCESS
	ERROR_NO_DEVICE     usbError = C.LIBUSB_ERROR_NO_DEVICE
	ERROR_NOT_FOUND     usbError = C.LIBUSB_ERROR_NOT_FOUND
	ERROR_BUSY          usbError = C.LIBUSB_ERROR_BUSY
	ERROR_TIMEOUT       usbError = C.LIBUSB_ERROR_TIMEOUT
	ERROR_OVERFLOW      usbError = C.LIBUSB_ERROR_OVERFLOW
	ERROR_PIPE          usbError = C.LIBUSB_ERROR_PIPE
	ERROR_INTERRUPTED   usbError = C.LIBUSB_ERROR_INTERRUPTED
	ERROR_NO_MEM        usbError = C.LIBUSB_ERROR_NO_MEM
	ERROR_NOT_SUPPORTED usbError = C.LIBUSB_ERROR_NOT_SUPPORTED
	ERROR_OTHER         usbError = C.LIBUSB_ERROR_OTHER
)

var usbErrorString = map[usbError]string{
	C.LIBUSB_SUCCESS:             "success",
	C.LIBUSB_ERROR_IO:            "i/o error",
	C.LIBUSB_ERROR_INVALID_PARAM: "invalid param",
	C.LIBUSB_ERROR_ACCESS:        "bad access",
	C.LIBUSB_ERROR_NO_DEVICE:     "no device",
	C.LIBUSB_ERROR_NOT_FOUND:     "not found",
	C.LIBUSB_ERROR_BUSY:          "device or resource busy",
	C.LIBUSB_ERROR_TIMEOUT:       "timeout",
	C.LIBUSB_ERROR_OVERFLOW:      "overflow",
	C.LIBUSB_ERROR_PIPE:          "pipe error",
	C.LIBUSB_ERROR_INTERRUPTED:   "interrupted",
	C.LIBUSB_ERROR_NO_MEM:        "out of memory",
	C.LIBUSB_ERROR_NOT_SUPPORTED: "not supported",
	C.LIBUSB_ERROR_OTHER:         "unknown error",
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
		var c_devlist **C.libusb_device
		cnt := C.libusb_get_device_list(nil, &c_devlist)
		defer C.libusb_free_device_list(c_devlist, 1)

		var dev_list []*C.libusb_device
		*(*reflect.SliceHeader)(unsafe.Pointer(&dev_list)) = reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(c_devlist)),
			Len:  int(cnt),
			Cap:  int(cnt),
		}

		for _, dev := range dev_list {
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
	//_, _, errorp := syscall.Syscall(syscall.SYS_IOCTL,
	//		uintptr(dev.handle),
	//		uintptr(C.makeHidIoCSFeature(C.int(len(data)))),
	//		uintptr(unsafe.Pointer(&data[0])))
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
