package devices

import (
	"errors"
	"strings"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/writer"
	"gitlab.com/gomidi/rtmididrv"

	// v2: gitlab.com/gomidi/midi/v2/drivers
	// _ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
	// v2: gitlab.com/gomidi/midi/v2

	"golang.org/x/exp/slog"
)

const (
	midiMaxRepeat = 50 // maximum number a message is repeated
)

var (
	ErrMIDIDeviceNotFound       = errors.New("MIDI Device not found")
	ErrMIDIDeviceNotInitialized = errors.New("MIDI Device not initialized")
)

// MidiController is the public interface to send out MIDI controller messages to a device
type MidiController interface {
	Open() error
	Close() error
	SendCommand(controller uint8, value uint8, repeat bool) error
}

// midiControllerCommand contains a single command that will be send out
type midiControllerCommand struct {
	controller uint8
	value      uint8
	repeat     bool
}

// midiControl contains all driver and channel variables in required for the communication
type midiControl struct {
	deviceName string
	delay      time.Duration
	channel    uint8
	driver     midi.Driver // v2: drivers.Driver
	output     midi.Out
	writer     *writer.Writer

	commandCh chan *midiControllerCommand
	quitCh    chan struct{}
}

// TODO: update to goMidi V2

// NewMIDIController creates a new MidiController instance with the specified parameters. If nil is passed as driver
// the default driver will be used (rtmididrv).
// delay specifies the time between each command message, in case the message should be send repeatedly.
func NewMIDIController(driver midi.Driver, deviceName string, delay time.Duration, channel uint8) MidiController {
	return &midiControl{driver: driver, deviceName: deviceName, delay: delay, channel: channel}
}

// GetMIDIDevices returns a list of all devices availalbe for the specified driver. If nil is passed as driver
// the default driver will be used (rtmididrv)
func GetMIDIDevices(driver midi.Driver) ([]string, error) {
	var drv midi.Driver
	var err error

	if driver == nil {
		drv, err = rtmididrv.New()
		if err != nil {
			return nil, err
		}
		defer drv.Close()
	} else {
		drv = driver
	}

	outs, err := drv.Outs() // v2: drivers.Outs()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(outs))
	for _, v := range outs {
		result = append(result, v.String())
	}
	return result, nil
}

// Open connects to the driver specified during instance creation, sets the channel used for the MIDI messages and starts
// the goroutine used for message sending
func (mc *midiControl) Open() error {
	if mc.driver == nil {
		drv, err := rtmididrv.New()
		if err != nil {
			slog.Error("rtmididrv: can't open new driver", err)
		}
		mc.driver = drv
	}

	outs, err := mc.driver.Outs()
	if err != nil {
		slog.Error("midi.driver: can't find out ports", err)
	}
	for i, v := range outs {
		// slog.Info(fmt.Sprint(i) + " " + v.String())
		if strings.Contains(v.String(), mc.deviceName) {
			mc.output = outs[i]
		}
	}

	if mc.output == nil {
		return ErrMIDIDeviceNotFound
	}

	if err := mc.output.Open(); err != nil {
		slog.Error("midi.port: can't open selected port", err)
	}

	mc.writer = writer.New(mc.output)
	mc.writer.SetChannel(mc.channel)

	mc.commandCh = make(chan *midiControllerCommand, 1)
	mc.quitCh = make(chan struct{})

	go mc.commandExecutor()
	return nil
}

// Close stops the goroutine and closes all channels and drivers
func (mc *midiControl) Close() error {
	if mc.quitCh != nil {
		close(mc.quitCh)
	}
	errout := mc.output.Close()
	errdrv := mc.driver.Close()
	if errout != nil {
		return errout
	} else if errdrv != nil {
		return errdrv
	}
	return nil
}

// SendCommand sends a ControllerChange MIDI command to the current MIDI device. If repeat is true then the message
// will be send up to midiMaxRepeat times with a delay as specified during instance creation
func (mc *midiControl) SendCommand(controller uint8, value uint8, repeat bool) error {
	if mc.output == nil {
		return ErrMIDIDeviceNotInitialized
	}
	cmd := &midiControllerCommand{controller: controller, value: value, repeat: repeat}
	mc.commandCh <- cmd
	return nil
}

// commandExecutor sends out MIDI messages received through the commandch channel. It also takes care of sending messages out
// repeatedly, in case it is requested
func (mc *midiControl) commandExecutor() {
	type tickStruct struct {
		counter int
		value   uint8
	}

	repeatcmd := make(map[uint8]tickStruct)
	tick := time.NewTicker(mc.delay)
	defer tick.Stop()

	tick.Stop()

	for {
		select {
		case <-mc.quitCh:
			return
		case cmd := <-mc.commandCh:
			//log.Printf("Controller: %v, Value: %v, Repeat: %v\n", cmd.controller, cmd.value, cmd.repeat)
			if cmd.value <= 127 {
				writer.ControlChange(mc.writer, cmd.controller, cmd.value)
			}
			if cmd.repeat {
				repeatcmd[cmd.controller] = tickStruct{counter: midiMaxRepeat, value: cmd.value}
				tick.Reset(mc.delay)
			} else {
				delete(repeatcmd, cmd.controller)
				if len(repeatcmd) == 0 {
					tick.Stop()
				}
			}
		case <-tick.C:
			for k, v := range repeatcmd {
				if v.counter > 1 {
					//log.Printf("Controller: %v, Value: %v, Repeat-Counter: %v\n", k, v.value, v.counter)
					writer.ControlChange(mc.writer, k, v.value)
					v.counter--
					repeatcmd[k] = v
				} else {
					delete(repeatcmd, k)
				}
			}
			if len(repeatcmd) == 0 {
				tick.Stop()
			}
		}
	}
}
