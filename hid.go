// HID package to access Human Interface Devices.
// The platform specific parts of this package are heavily based on
// Signal 11 - HIDAPI. (https://github.com/signal11/hidapi)
package hid

import "strings"

// DeviceInfo provides general information about a device
type DeviceInfo struct {
	// Path contains a Platform-specific device path which is used to identify the device
	Path string
	// VendorId contains the USB Vendor ID of the device
	VendorId uint16
	// ProductId contains the USB Product ID of the device
	ProductId uint16
	// VersionNumber contains the Version / Release Number of the device
	VersionNumber uint16
	// Manufacturer of the USB device
	Manufacturer string
	// Product contains the product name of the device
	Product string

	InputReportLength   uint16
	OutputReportLength  uint16
	FeatureReportLength uint16
}

// Device interface for an opened HID USB device
type Device interface {
	// Close closes the device and release all keept resources.
	Close()
	// Write to the device
	// (technically a HID report with type 'output' is send to the device)
	Write([]byte) error
	// Write to the device
	// (technically a HID report with type 'feature' is send to the device)
	WriteFeature([]byte) error
	// Preform an interrupt transfer to the device
	WriteInterrupt(byte, []byte) (int, error)
}

// FindDevices iterates through all devices with a given vendor and product id
func FindDevices(vendor uint16, product uint16) <-chan *DeviceInfo {
	result := make(chan *DeviceInfo)
	go func() {
		for dev := range Devices() {
			if dev.VendorId == vendor && dev.ProductId == product {
				result <- dev
			}
		}
		close(result)
	}()
	return result
}

// FindDevicesByProduct iterates through all devices with a given vendor and product id
func FindDevicesByProduct(product string) <-chan *DeviceInfo {
	result := make(chan *DeviceInfo)

	go func() {
		for dev := range Devices() {
			if strings.Contains(dev.Product, product) {
				result <- dev
			}
		}
		close(result)
	}()

	return result
}
