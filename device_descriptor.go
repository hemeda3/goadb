package adb

import "fmt"

//go:generate stringer -type=deviceDescriptorType
type deviceDescriptorType int

const (
	// host:transport-any and host:<request>
	DeviceAny deviceDescriptorType = iota
	// host:transport:<Serial> and host-Serial:<Serial>:<request>
	DeviceSerial
	// host:transport-usb and host-usb:<request>
	DeviceUsb
	// host:transport-local and host-local:<request>
	DeviceLocal
)

type DeviceDescriptor struct {
	DescriptorType deviceDescriptorType

	// Only used if Type is DeviceSerial.
	Serial string
}

func AnyDevice() DeviceDescriptor {
	return DeviceDescriptor{DescriptorType: DeviceAny}
}

func AnyUsbDevice() DeviceDescriptor {
	return DeviceDescriptor{DescriptorType: DeviceUsb}
}

func AnyLocalDevice() DeviceDescriptor {
	return DeviceDescriptor{DescriptorType: DeviceLocal}
}

func DeviceWithSerial(serial string) DeviceDescriptor {
	return DeviceDescriptor{
		DescriptorType: DeviceSerial,
		Serial:         serial,
	}
}

func (d DeviceDescriptor) String() string {
	if d.DescriptorType == DeviceSerial {
		return fmt.Sprintf("%s[%s]", d.DescriptorType, d.Serial)
	}
	return d.DescriptorType.String()
}

func (d DeviceDescriptor) getHostPrefix() string {
	switch d.DescriptorType {
	case DeviceAny:
		return "host"
	case DeviceUsb:
		return "host-usb"
	case DeviceLocal:
		return "host-local"
	case DeviceSerial:
		return fmt.Sprintf("host-Serial:%s", d.Serial)
	default:
		panic(fmt.Sprintf("invalid DeviceDescriptorType: %v", d.DescriptorType))
	}
}

func (d DeviceDescriptor) getTransportDescriptor() string {
	switch d.DescriptorType {
	case DeviceAny:
		return "transport-any"
	case DeviceUsb:
		return "transport-usb"
	case DeviceLocal:
		return "transport-local"
	case DeviceSerial:
		return fmt.Sprintf("transport:%s", d.Serial)
	default:
		panic(fmt.Sprintf("invalid DeviceDescriptorType: %v", d.DescriptorType))
	}
}
