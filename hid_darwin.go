package hid

/*
#cgo LDFLAGS: -L . -L/usr/local/lib -framework CoreFoundation -framework IOKit

#include <IOKit/hid/IOHIDManager.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>


void deviceUnpluged(IOHIDDeviceRef osd, IOReturn ret, void* dev);
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

func ioReturnToErr(ret C.IOReturn) error {
	switch ret {
	case C.kIOReturnSuccess:
		return nil
	case C.kIOReturnError:
		return errors.New("general error ")
	case C.kIOReturnNoMemory:
		return errors.New("can't allocate memory")
	case C.kIOReturnNoResources:
		return errors.New("resource shortage")
	case C.kIOReturnIPCError:
		return errors.New("error during IPC")
	case C.kIOReturnNoDevice:
		return errors.New("no such device")
	case C.kIOReturnNotPrivileged:
		return errors.New("privilege violation")
	case C.kIOReturnBadArgument:
		return errors.New("invalid argument")
	case C.kIOReturnLockedRead:
		return errors.New("device read locked")
	case C.kIOReturnLockedWrite:
		return errors.New("device write locked")
	case C.kIOReturnExclusiveAccess:
		return errors.New("exclusive access and device already open")
	case C.kIOReturnBadMessageID:
		return errors.New("sent/received messages had different msg_id")
	case C.kIOReturnUnsupported:
		return errors.New("unsupported function")
	case C.kIOReturnVMError:
		return errors.New("misc. VM failure")
	case C.kIOReturnInternalError:
		return errors.New("internal error")
	case C.kIOReturnIOError:
		return errors.New("General I/O error")
	case C.kIOReturnCannotLock:
		return errors.New("can't acquire lock")
	case C.kIOReturnNotOpen:
		return errors.New("device not open")
	case C.kIOReturnNotReadable:
		return errors.New("read not supported")
	case C.kIOReturnNotWritable:
		return errors.New("write not supported")
	case C.kIOReturnNotAligned:
		return errors.New("alignment error")
	case C.kIOReturnBadMedia:
		return errors.New("Media Error")
	case C.kIOReturnStillOpen:
		return errors.New("device(s) still open")
	case C.kIOReturnRLDError:
		return errors.New("rld failure")
	case C.kIOReturnDMAError:
		return errors.New("DMA failure")
	case C.kIOReturnBusy:
		return errors.New("Device Busy")
	case C.kIOReturnTimeout:
		return errors.New("I/O Timeout")
	case C.kIOReturnOffline:
		return errors.New("device offline")
	case C.kIOReturnNotReady:
		return errors.New("not ready")
	case C.kIOReturnNotAttached:
		return errors.New("device not attached")
	case C.kIOReturnNoChannels:
		return errors.New("no DMA channels left")
	case C.kIOReturnNoSpace:
		return errors.New("no space for data")
	case C.kIOReturnPortExists:
		return errors.New("port already exists")
	case C.kIOReturnCannotWire:
		return errors.New("can't wire down physical memory")
	case C.kIOReturnNoInterrupt:
		return errors.New("no interrupt attached")
	case C.kIOReturnNoFrames:
		return errors.New("no DMA frames enqueued")
	case C.kIOReturnMessageTooLarge:
		return errors.New("oversized msg received on interrupt port")
	case C.kIOReturnNotPermitted:
		return errors.New("not permitted")
	case C.kIOReturnNoPower:
		return errors.New("no power to device")
	case C.kIOReturnNoMedia:
		return errors.New("media not present")
	case C.kIOReturnUnformattedMedia:
		return errors.New("edia not formatted")
	case C.kIOReturnUnsupportedMode:
		return errors.New("no such mode")
	case C.kIOReturnUnderrun:
		return errors.New("data underrun")
	case C.kIOReturnOverrun:
		return errors.New("data overrun")
	case C.kIOReturnDeviceError:
		return errors.New("the device is not working properly!")
	case C.kIOReturnNoCompletion:
		return errors.New("a completion routine is required")
	case C.kIOReturnAborted:
		return errors.New("operation aborted")
	case C.kIOReturnNoBandwidth:
		return errors.New("bus bandwidth would be exceeded")
	case C.kIOReturnNotResponding:
		return errors.New("device not responding")
	case C.kIOReturnIsoTooOld:
		return errors.New("isochronous I/O request for distant past!")
	case C.kIOReturnIsoTooNew:
		return errors.New("isochronous I/O request for distant future")
	case C.kIOReturnNotFound:
		return errors.New("data was not found")
	}
	return errors.New("Unknown error")
}

type cleanupDeviceManagerFn func()
type osxDevice struct {
	osDevice     C.IOHIDDeviceRef
	disconnected bool
	closeDM      cleanupDeviceManagerFn
}

func cfstring(s string) C.CFStringRef {
	bytes := C.CString(s)
	defer C.free(unsafe.Pointer(bytes))
	return C.CFStringCreateWithBytes(C.kCFAllocatorDefault, (*C.UInt8)(unsafe.Pointer(bytes)), C.CFIndex(len(s)), C.kCFStringEncodingUTF8, 0)
}

func releaseCFString(s C.CFStringRef) {
	if s != 0 {
		C.CFRelease(C.CFTypeRef(s))
	}
}

func gostring(cfs C.CFStringRef) string {
	if cfs == 0 {
		return ""
	}

	length := C.CFStringGetLength(cfs)
	if length == 0 {
		return ""
	}
	rng := C.CFRange{location: 0, length: length}
	var usedBufLen C.CFIndex
	if C.CFStringGetBytes(cfs, rng, C.kCFStringEncodingUTF8, 0, 0, nil, 0, &usedBufLen) != length || usedBufLen == 0 {
		return ""
	}
	buf := make([]byte, int(usedBufLen))
	if C.CFStringGetBytes(cfs, rng, C.kCFStringEncodingUTF8, 0, 0, (*C.UInt8)(unsafe.Pointer(&buf[0])), C.CFIndex(len(buf)), &usedBufLen) != length {
		return ""
	}
	return string(buf[:int(usedBufLen)])
}

func getIntProp(device C.IOHIDDeviceRef, key string) int32 {
	cfKey := cfstring(key)
	defer releaseCFString(cfKey)
	var value int32

	ref := C.IOHIDDeviceGetProperty(device, cfKey)
	if ref != 0 {
		if C.CFGetTypeID(ref) == C.CFNumberGetTypeID() {
			C.CFNumberGetValue(C.CFNumberRef(ref), C.kCFNumberSInt32Type, unsafe.Pointer(&value))
			return value
		}
	}
	return 0
}

func getStringProp(device C.IOHIDDeviceRef, key string) string {
	cfKey := cfstring(key)
	defer releaseCFString(cfKey)
	s := C.IOHIDDeviceGetProperty(device, cfKey)
	return gostring(C.CFStringRef(s))
}

func getPath(osDev C.IOHIDDeviceRef) string {
	return fmt.Sprintf("%s_%04x_%04x_%08x",
		getStringProp(osDev, C.kIOHIDTransportKey),
		uint16(getIntProp(osDev, C.kIOHIDVendorIDKey)),
		uint16(getIntProp(osDev, C.kIOHIDProductIDKey)),
		uint32(getIntProp(osDev, C.kIOHIDLocationIDKey)))
}

func iterateDevices(action func(device C.IOHIDDeviceRef) bool) cleanupDeviceManagerFn {
	mgr := C.IOHIDManagerCreate(C.kCFAllocatorDefault, C.kIOHIDOptionsTypeNone)
	C.IOHIDManagerSetDeviceMatching(mgr, 0)
	C.IOHIDManagerOpen(mgr, C.kIOHIDOptionsTypeNone)

	allDevicesSet := C.IOHIDManagerCopyDevices(mgr)
	defer C.CFRelease(C.CFTypeRef(allDevicesSet))
	devCnt := C.CFSetGetCount(allDevicesSet)
	allDevices := make([]unsafe.Pointer, uint64(devCnt))
	C.CFSetGetValues(allDevicesSet, &allDevices[0])

	for _, pDev := range allDevices {
		if !action(C.IOHIDDeviceRef(pDev)) {
			break
		}
	}
	return func() {
		C.IOHIDManagerClose(mgr, C.kIOHIDOptionsTypeNone)
		C.CFRelease(C.CFTypeRef(mgr))
	}
}

func Devices() <-chan *DeviceInfo {
	result := make(chan *DeviceInfo)
	go func() {
		iterateDevices(func(device C.IOHIDDeviceRef) bool {
			result <- &DeviceInfo{
				VendorId:            uint16(getIntProp(device, C.kIOHIDVendorIDKey)),
				ProductId:           uint16(getIntProp(device, C.kIOHIDProductIDKey)),
				VersionNumber:       uint16(getIntProp(device, C.kIOHIDVersionNumberKey)),
				Manufacturer:        getStringProp(device, C.kIOHIDManufacturerKey),
				Product:             getStringProp(device, C.kIOHIDProductKey),
				InputReportLength:   uint16(getIntProp(device, C.kIOHIDMaxInputReportSizeKey)),
				OutputReportLength:  uint16(getIntProp(device, C.kIOHIDMaxOutputReportSizeKey)),
				FeatureReportLength: uint16(getIntProp(device, C.kIOHIDMaxFeatureReportSizeKey)),
				Path:                getPath(device),
			}
			return true
		})()
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
	err := errors.New("device not found")
	var dev *osxDevice
	closeDM := iterateDevices(func(device C.IOHIDDeviceRef) bool {
		if getPath(device) == di.Path {
			res := C.IOHIDDeviceOpen(device, C.kIOHIDOptionsTypeSeizeDevice)
			if res == C.kIOReturnSuccess {
				C.CFRetain(C.CFTypeRef(device))
				dev = &osxDevice{osDevice: device}
				err = nil
				C.IOHIDDeviceRegisterRemovalCallback(device, (C.IOHIDCallback)(unsafe.Pointer(C.deviceUnpluged)), unsafe.Pointer(dev))
			} else {
				err = ioReturnToErr(res)
			}
			return false
		}
		return true
	})
	if dev != nil {
		dev.closeDM = closeDM
	}

	return dev, err
}

//export deviceUnpluged
func deviceUnpluged(osdev C.IOHIDDeviceRef, result C.IOReturn, dev unsafe.Pointer) {
	od := (*osxDevice)(dev)
	od.disconnected = true
	od.Close()
}

func (dev *osxDevice) Close() {
	if !dev.disconnected {
		C.IOHIDDeviceClose(dev.osDevice, C.kIOHIDOptionsTypeSeizeDevice)
		dev.disconnected = true
	}
	if dev.osDevice != 0 {
		C.CFRelease(C.CFTypeRef(dev.osDevice))
		dev.osDevice = 0
	}
	if dev.closeDM != nil {
		dev.closeDM()
		dev.closeDM = nil
	}
}

func (dev *osxDevice) setReport(typ C.IOHIDReportType, data []byte) error {
	var reportNo int32 = int32(data[0])
	if reportNo == 0 {
		data = data[1:]
	}

	if !dev.disconnected {
		res := C.IOHIDDeviceSetReport(dev.osDevice, typ, C.CFIndex(reportNo), (*C.uint8_t)(&data[0]), C.CFIndex(len(data)))

		if res == C.kIOReturnSuccess {
			return nil
		} else {
			return ioReturnToErr(res)
		}
	}

	return errors.New("device disconnected")
}

func (dev *osxDevice) WriteFeature(data []byte) error {
	return dev.setReport(C.kIOHIDReportTypeFeature, data)
}

func (dev *osxDevice) Write(data []byte) error {
	return dev.setReport(C.kIOHIDReportTypeOutput, data)
}

func (dev *osxDevice) WriteInterrupt(endpoint byte, data []byte) (int, error) {
	return 0, errors.New("WriteInterrupt is not implemented")
}
