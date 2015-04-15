package hid

/*

#include <linux/hidraw.h>
#include <libudev.h>
#include <sys/ioctl.h>
#include <locale.h>

static inline int makeHidIoCSFeature(int len) {
  return HIDIOCSFEATURE(len)
}

*/

import "C"

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func init() {
	C.setlocale(LC_CTYPE, nil)
}

type linuxDevice struct {
	handle int
}

func Devices() <-chan *DeviceInfo {
	result := make(chan *DeviceInfo)
	go func() {
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
	_, _, errorp := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(dev.handle),
		uintptr(C.makeHidIoCSFeature(len(data))),
		uintptr(unsafe.Pointer(&data[0])))
	return os.NewSyscallError("ioctl", errorp)
}

func (dev *linuxDevice) Write(data []byte) error {
	n, err := syscall.Write(dev.handle, data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("written bytes missmatch!")
	}
}
