package hid

// #cgo pkg-config: libusb-1.0
// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import "fmt"

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

	// Request types to use in ControlTransfer().
	REQUEST_TYPE_STANDARD = (0x00 << 5)
	REQUEST_TYPE_CLASS    = (0x01 << 5)
	REQUEST_TYPE_VENDOR   = (0x02 << 5)
	REQUEST_TYPE_RESERVED = (0x03 << 5)

	// Recipient bits for the reqType of ControlTransfer(). Values 4 - 31
	// are reserved.
	RECIPIENT_DEVICE    = 0x00
	RECIPIENT_INTERFACE = 0x01
	RECIPIENT_ENDPOINT  = 0x02
	RECIPIENT_OTHER     = 0x03

	// in: device-to-host
	ENDPOINT_IN = 0x80

	// out: host-to-device
	ENDPOINT_OUT = 0x00

	// Descriptor types as defined by the USB specification.
	DT_DEVICE    = 0x01
	DT_CONFIG    = 0x02
	DT_STRING    = 0x03
	DT_INTERFACE = 0x04
	DT_ENDPOINT  = 0x05
	DT_HID       = 0x21
	DT_REPORT    = 0x22
	DT_PHYSICAL  = 0x23
	DT_HUB       = 0x29

	// Standard request types, as defined in table 9-3 of the USB2 specifications
	REQUEST_GET_STATUS    = 0x00
	REQUEST_CLEAR_FEATURE = 0x01

	// Set or enable a specific feature
	REQUEST_SET_FEATURE = 0x03

	// Set device address for all future accesses
	REQUEST_SET_ADDRESS = 0x05

	// Get the specified descriptor
	REQUEST_GET_DESCRIPTOR = 0x06

	// Used to update existing descriptors or add new descriptors
	REQUEST_SET_DESCRIPTOR = 0x07

	// Get the current device configuration value
	REQUEST_GET_CONFIGURATION = 0x08

	// Set device configuration
	REQUEST_SET_CONFIGURATION = 0x09

	// Return the selected alternate setting for the specified interface.
	REQUEST_GET_INTERFACE = 0x0A

	// Select an alternate interface for the specified interface
	REQUEST_SET_INTERFACE = 0x0B

	// Set then report an endpoint's synchronization frame
	REQUEST_SYNCH_FRAME = 0x0C
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
