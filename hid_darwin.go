package hid

/*
#cgo LDFLAGS: -L . -L/usr/local/lib -framework CoreFoundation -framework IOKit -fconstant-cfstrings

#include <IOKit/hid/IOHIDManager.h>
#include <CoreFoundation/CoreFoundation.h>


static inline CFIndex cfstring_utf8_length(CFStringRef str, CFIndex *need) {
  CFIndex n, usedBufLen;
  CFRange rng = CFRangeMake(0, CFStringGetLength(str));

  return CFStringGetBytes(str, rng, kCFStringEncodingUTF8, 0, 0, NULL, 0, need);
}

void deviceUnpluged(IOHIDDeviceRef osd, IOReturn ret, void* dev);
*/
import "C"

import (
	"errors"
	"fmt"
	"reflect"
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
	n := C.CFIndex(len(s))
	return C.CFStringCreateWithBytes(nil, *(**C.UInt8)(unsafe.Pointer(&s)), n, C.kCFStringEncodingUTF8, 0)
}

func gostring(cfs C.CFStringRef) string {
	if cfs == nil {
		return ""
	}

	var usedBufLen C.CFIndex
	n := C.cfstring_utf8_length(cfs, &usedBufLen)
	if n <= 0 {
		return ""
	}
	rng := C.CFRange{location: C.CFIndex(0), length: n}
	buf := make([]byte, int(usedBufLen))

	bufp := unsafe.Pointer(&buf[0])
	C.CFStringGetBytes(cfs, rng, C.kCFStringEncodingUTF8, 0, 0, (*C.UInt8)(bufp), C.CFIndex(len(buf)), &usedBufLen)

	sh := &reflect.StringHeader{
		Data: uintptr(bufp),
		Len:  int(usedBufLen),
	}
	return *(*string)(unsafe.Pointer(sh))
}

func getIntProp(device C.IOHIDDeviceRef, key C.CFStringRef) int32 {
	var value int32

	ref := C.IOHIDDeviceGetProperty(device, key)
	if ref != nil {
		if C.CFGetTypeID(ref) == C.CFNumberGetTypeID() {
			C.CFNumberGetValue(C.CFNumberRef(ref), C.kCFNumberSInt32Type, unsafe.Pointer(&value))
			return value
		}
	}
	return 0
}

func getStringProp(device C.IOHIDDeviceRef, key C.CFStringRef) string {
	s := C.IOHIDDeviceGetProperty(device, key)
	return gostring(C.CFStringRef(s))
}

func getPath(osDev C.IOHIDDeviceRef) string {
	return fmt.Sprintf("%s_%04x_%04x_%08x",
		getStringProp(osDev, cfstring(C.kIOHIDTransportKey)),
		uint16(getIntProp(osDev, cfstring(C.kIOHIDVendorIDKey))),
		uint16(getIntProp(osDev, cfstring(C.kIOHIDProductIDKey))),
		uint32(getIntProp(osDev, cfstring(C.kIOHIDLocationIDKey))))
}

func iterateDevices(action func(device C.IOHIDDeviceRef) bool) cleanupDeviceManagerFn {
	mgr := C.IOHIDManagerCreate(C.kCFAllocatorDefault, C.kIOHIDOptionsTypeNone)
	C.IOHIDManagerSetDeviceMatching(mgr, nil)
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
				VendorId:            uint16(getIntProp(device, cfstring(C.kIOHIDVendorIDKey))),
				ProductId:           uint16(getIntProp(device, cfstring(C.kIOHIDProductIDKey))),
				VersionNumber:       uint16(getIntProp(device, cfstring(C.kIOHIDVersionNumberKey))),
				Manufacturer:        getStringProp(device, cfstring(C.kIOHIDManufacturerKey)),
				Product:             getStringProp(device, cfstring(C.kIOHIDProductKey)),
				InputReportLength:   uint16(getIntProp(device, cfstring(C.kIOHIDMaxInputReportSizeKey))),
				OutputReportLength:  uint16(getIntProp(device, cfstring(C.kIOHIDMaxOutputReportSizeKey))),
				FeatureReportLength: uint16(getIntProp(device, cfstring(C.kIOHIDMaxFeatureReportSizeKey))),
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
	if dev.osDevice != nil {
		C.CFRelease(C.CFTypeRef(dev.osDevice))
		dev.osDevice = nil
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

func (dev *osxDevice) getReport(reportNo int32) ([]byte, error) {
	var bufSize C.CFIndex = (C.CFIndex)(getIntProp(dev.osDevice, cfstring(C.kIOHIDMaxInputReportSizeKey)))
	report := make([]byte, bufSize)

	if !dev.disconnected {
		res := C.IOHIDDeviceGetReport(dev.osDevice, C.kIOHIDReportTypeInput, C.CFIndex(reportNo), (*C.uint8_t)(&report[0]), &bufSize)
		if res == C.kIOReturnSuccess {
			return report[:bufSize], nil
		} else {
			return nil, ioReturnToErr(res)
		}
	}

	return nil, errors.New("device disconnected")
}

func (dev *osxDevice) Read(reportNo byte) ([]byte, error) {
	return dev.getReport(int32(reportNo))
}
