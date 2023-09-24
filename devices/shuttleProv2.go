package devices

import (
	"errors"

	//"github.com/bearsh/hid"
	"github.com/awitez/shuttleMidi/hid" // local
)

// USB HID device information
const (
	shuttleProV2VendorId  = 0x0b33
	shuttleProV2ProductId = 0x0030
)

var (
	ErrShuttleProV2DeviceNotFound  = errors.New("no ShuttlePRO v2 found")
	ErrShuttleProV2DeviceNotOpened = errors.New("ShuttlePRO v2: No device opened")
)

// ShuttleStatus contains a event channel for all ShuttleProv2 hardware controls.
// The channels have to be created by the consuming module.
type ShuttleProV2Status struct {
	WheelPosition chan int8
	wheelValue    int8

	DialDirection chan int8
	dialValue     uint8

	Button1Pressed  chan bool
	Button2Pressed  chan bool
	Button3Pressed  chan bool
	Button4Pressed  chan bool
	Button5Pressed  chan bool
	Button6Pressed  chan bool
	Button7Pressed  chan bool
	Button8Pressed  chan bool
	Button9Pressed  chan bool
	Button10Pressed chan bool
	Button11Pressed chan bool
	Button12Pressed chan bool
	Button13Pressed chan bool
	Button14Pressed chan bool
	Button15Pressed chan bool

	button1Value  bool
	button2Value  bool
	button3Value  bool
	button4Value  bool
	button5Value  bool
	button6Value  bool
	button7Value  bool
	button8Value  bool
	button9Value  bool
	button10Value bool
	button11Value bool
	button12Value bool
	button13Value bool
	button14Value bool
	button15Value bool
}

type ShuttleProV2 struct {
	devHandle *hid.Device
	devInfo   hid.DeviceInfo
	err       error

	ShuttleProV2Status
}

// readDevice is a goroutine and continously reads the device status and sends out events through the channels part of ShuttleStatus
func (shuttlePro *ShuttleProV2) readDevice() {

	if shuttlePro.devHandle == nil {
		shuttlePro.err = ErrShuttleProV2DeviceNotOpened
		return
	}
	shuttlePro.devHandle.SetNonblocking(false)

	for {
		var buf = make([]byte, 48)                                // a slice is always a pointer
		if _, err := shuttlePro.devHandle.Read(buf); err != nil { // can't read from HID
			shuttlePro.err = err
			return
		}

		wheelPos := int8(buf[0])
		dialPos := uint8(buf[1])

		// see ShuttleProV2rawUSBdata.txt
		b1_pressed := buf[3]&(1<<0) > 0
		b2_pressed := buf[3]&(1<<1) > 0
		b3_pressed := buf[3]&(1<<2) > 0
		b4_pressed := buf[3]&(1<<3) > 0
		b5_pressed := buf[3]&(1<<4) > 0
		b6_pressed := buf[3]&(1<<5) > 0
		b7_pressed := buf[3]&(1<<6) > 0
		b8_pressed := buf[3]&(1<<7) > 0
		b9_pressed := buf[4]&(1<<0) > 0
		b10_pressed := buf[4]&(1<<1) > 0
		b11_pressed := buf[4]&(1<<2) > 0
		b12_pressed := buf[4]&(1<<3) > 0
		b13_pressed := buf[4]&(1<<4) > 0
		b14_pressed := buf[4]&(1<<5) > 0
		b15_pressed := buf[4]&(1<<6) > 0

		if wheelPos != shuttlePro.wheelValue && shuttlePro.WheelPosition != nil { // wheel was moved
			shuttlePro.WheelPosition <- wheelPos
			shuttlePro.wheelValue = wheelPos
		}
		if dialPos != shuttlePro.dialValue && shuttlePro.DialDirection != nil { // dial was moved
			dial_delta := int8(dialPos - shuttlePro.dialValue)
			if dial_delta == 1 || dial_delta == -1 { // only use if difference is a single step. Else it's the first read
				shuttlePro.DialDirection <- dial_delta
			}
			shuttlePro.dialValue = dialPos
		}
		if b1_pressed != shuttlePro.button1Value {
			shuttlePro.Button1Pressed <- b1_pressed
			shuttlePro.button1Value = b1_pressed
		}
		if b2_pressed != shuttlePro.button2Value {
			shuttlePro.Button2Pressed <- b2_pressed
			shuttlePro.button2Value = b2_pressed
		}
		if b3_pressed != shuttlePro.button3Value {
			shuttlePro.Button3Pressed <- b3_pressed
			shuttlePro.button3Value = b3_pressed
		}
		if b4_pressed != shuttlePro.button4Value {
			shuttlePro.Button4Pressed <- b4_pressed
			shuttlePro.button4Value = b4_pressed
		}
		if b5_pressed != shuttlePro.button5Value {
			shuttlePro.Button5Pressed <- b5_pressed
			shuttlePro.button5Value = b5_pressed
		}
		if b6_pressed != shuttlePro.button6Value {
			shuttlePro.Button6Pressed <- b6_pressed
			shuttlePro.button6Value = b6_pressed
		}
		if b7_pressed != shuttlePro.button7Value {
			shuttlePro.Button7Pressed <- b7_pressed
			shuttlePro.button7Value = b7_pressed
		}
		if b8_pressed != shuttlePro.button8Value {
			shuttlePro.Button8Pressed <- b8_pressed
			shuttlePro.button8Value = b8_pressed
		}
		if b9_pressed != shuttlePro.button9Value {
			shuttlePro.Button9Pressed <- b9_pressed
			shuttlePro.button9Value = b9_pressed
		}
		if b10_pressed != shuttlePro.button10Value {
			shuttlePro.Button10Pressed <- b10_pressed
			shuttlePro.button10Value = b10_pressed
		}
		if b11_pressed != shuttlePro.button11Value {
			shuttlePro.Button11Pressed <- b11_pressed
			shuttlePro.button11Value = b11_pressed
		}
		if b12_pressed != shuttlePro.button12Value {
			shuttlePro.Button12Pressed <- b12_pressed
			shuttlePro.button12Value = b12_pressed
		}
		if b13_pressed != shuttlePro.button13Value {
			shuttlePro.Button13Pressed <- b13_pressed
			shuttlePro.button13Value = b13_pressed
		}
		if b14_pressed != shuttlePro.button14Value {
			shuttlePro.Button14Pressed <- b14_pressed
			shuttlePro.button14Value = b14_pressed
		}
		if b15_pressed != shuttlePro.button15Value {
			shuttlePro.Button15Pressed <- b15_pressed
			shuttlePro.button15Value = b15_pressed
		}
	}
}

// NewShuttleProV2 searches for available ShuttleProv2 devices and opens the first one it finds
func NewShuttleProV2() (*ShuttleProV2, error) {
	deviceInfo := hid.Enumerate(shuttleProV2VendorId, shuttleProV2ProductId)
	if len(deviceInfo) == 0 {
		return nil, ErrShuttleProV2DeviceNotFound
	}

	dev, err := deviceInfo[0].Open()
	if err != nil { // unable to open first ShuttleProV2
		return nil, err
	}

	status := ShuttleProV2Status{}
	sp := &ShuttleProV2{
		devHandle:          dev,
		devInfo:            deviceInfo[0],
		err:                nil,
		ShuttleProV2Status: status,
	}
	go sp.readDevice()
	return sp, nil
}
