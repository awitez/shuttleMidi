// ShuttleMidi sends MIDI events for the Contour ShuttleProV2

package main

import (
	"strings"
	"time"

	"github.com/awitez/shuttleMidi/devices"
	"github.com/awitez/shuttleMidi/icon"

	"log/slog"

	"fyne.io/systray"
	"github.com/gen2brain/dlgs"
	"github.com/spf13/viper"
)

// quitCh is the channel used to stop the goroutine handling the ShuttlePro events
var (
	quitCh   chan struct{}
	mControl midiController
)

// initSettings initializes the settings engine Viper. If it doesn't exist it is automatically created using the defaults
func initSettings() error {
	for k, v := range configDefaults {
		viper.SetDefault(k, v)
	}
	viper.SetConfigName("shuttleMidi")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/")

	if err := viper.ReadInConfig(); err != nil { // can't read
		slog.Error("viper: cannot read configfile", err)
		if _, ok := err.(viper.ConfigFileNotFoundError); ok { // file not found
			slog.Error("viper: config file not found", err)
			if err = viper.SafeWriteConfig(); err != nil { // can't write
				slog.Error("viper: can't save config", err)
				return err
			}
		} else { // file found but a different error
			return err
		}
		slog.Info("viper: config file was written")
	}
	return nil
}

// refreshDisplay transmits all values to the display device (Mackie Control)
func refreshDisplay(device string) { // TODO:  make solo state blink
	if !viper.GetBool("useDisplay") {
		return
	}
	textUpperRow := ""
	for i := range csPro {
		if csPro[i].latch {
			if csPro[i].state { // button is 'on'
				if (csPro[i].LCDchannel != 0) && (csPro[i].LCDrow == upperRow) {
					textUpperRow = textUpperRow + csPro[i].msgOn
				}
			} else { // button is 'off'
				if (csPro[i].LCDchannel != 0) && (csPro[i].LCDrow == upperRow) {
					textUpperRow = textUpperRow + csPro[i].msgOff
				}
			}
		}
	}

	textLowerRow := mainVolTable[uint8(mainVolume)]
	if csPro[LFEbutton].state {
		textLowerRow = textLowerRow + csPro[LFEbutton].msgOn
	} else {
		textLowerRow = textLowerRow + csPro[LFEbutton].msgOff
	}
	textLowerRow = textLowerRow + headPhoneVolTable[uint8(headPhoneVolume)]

	devices.DisplayLCDtext(device, csPro[LRbutton].LCDchannel, upperRow, textUpperRow)
	time.Sleep(displayRowsDelay * time.Millisecond) // wait a little bit, MCU device might be 'overwhelmed'
	devices.DisplayLCDtext(device, csPro[LRbutton].LCDchannel, lowerRow, textLowerRow)
}

// onReady is called by systray once the system tray menu can be created. It inializes the menu and opens the ShuttlePro device
func onReady() {
	shuttlePro, err := devices.NewShuttleProV2()
	if err != nil {
		if err == devices.ErrShuttleProV2DeviceNotFound {
			dlgs.Error(applicationName, "No ShuttlePro v2 device connected to this computer. Cannot continue.")
		} else {
			dlgs.Error(applicationName, err.Error())
		}
		slog.Error("devices: can't open ShuttleProv2", err)
		systray.Quit()
	}

	MIDIdevices, err := getMIDIDevices(nil)
	if err != nil {
		slog.Error("devices: can't get MIDIdevices", err)
	}

	controlName := viper.GetString("controlMidiDevice")
	displayName := viper.GetString("displayMidiDevice")
	menuExit := make(chan struct{})

	// build systray menues
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTooltip(applicationName)

	mControlMIDIMenu := systray.AddMenuItem("Control MIDI device", "")
	mDisplayMIDIMenu := systray.AddMenuItem("Display MIDI device", "")

	mControlMIDISubItems := make([]*systray.MenuItem, 0, len(MIDIdevices))
	mDisplayMIDISubItems := make([]*systray.MenuItem, 0, len(MIDIdevices))

	systray.AddSeparator()
	mReconnectShuttle := systray.AddMenuItem("Reconnect Shuttle", "")
	mRefreshDisplayItem := systray.AddMenuItem("Refresh Display", "")
	mUseDisplayItem := systray.AddMenuItemCheckbox("Use Display", "", viper.GetBool("useDisplay"))

	systray.AddSeparator()
	mUseMediaKeys := systray.AddMenuItemCheckbox("Control Music.app", "", viper.GetBool("useMediaKeys"))

	systray.AddSeparator()
	mQuitItem := systray.AddMenuItem("Quit", "")
	mQuitItem.Enable()

	for _, v := range MIDIdevices {
		mControlMIDISubItem := mControlMIDIMenu.AddSubMenuItemCheckbox(v, "", strings.Contains(v, controlName))
		mDisplayMIDISubItem := mDisplayMIDIMenu.AddSubMenuItemCheckbox(v, "", strings.Contains(v, displayName))

		mControlMIDISubItems = append(mControlMIDISubItems, mControlMIDISubItem)
		mDisplayMIDISubItems = append(mDisplayMIDISubItems, mDisplayMIDISubItem)
		title := v

		// a go routine for every sub menu item (MIDI device)
		go func() {
			for {
				select {
				case <-mControlMIDISubItem.ClickedCh:
					for _, v := range mControlMIDISubItems {
						v.Uncheck()
					}
					mControlMIDISubItem.Check()
					viper.Set("controlMidiDevice", title)
					viper.WriteConfig()
					startListeners(title, shuttlePro)
				case <-mDisplayMIDISubItem.ClickedCh:
					for _, v := range mDisplayMIDISubItems {
						v.Uncheck()
					}
					mDisplayMIDISubItem.Check()
					viper.Set("displayMidiDevice", title)
					viper.WriteConfig()
				case <-menuExit:
					return
				}
			}
		}()
	}

	go func() { // loop for menu items: 'Refresh Display' + 'Use Display' + 'Control Music.app'
		for {
			select {
			case <-mReconnectShuttle.ClickedCh:
				slog.Info("trying to reconnect ShuttlePro")
				err = devices.ReOpenShuttleProV2(shuttlePro)
				if err != nil {
					dlgs.Error(applicationName, err.Error())
				}
			case <-mRefreshDisplayItem.ClickedCh:
				refreshDisplay(viper.GetString("displayMidiDevice"))
			case <-mUseDisplayItem.ClickedCh:
				if mUseDisplayItem.Checked() {
					mUseDisplayItem.Uncheck()
					viper.Set("useDisplay", false)
					viper.WriteConfig()
					devices.ClearDisplay(viper.GetString("displayMidiDevice"), csPro[LRbutton].LCDchannel)
				} else {
					mUseDisplayItem.Check()
					viper.Set("useDisplay", true)
					viper.WriteConfig()
					refreshDisplay(viper.GetString("displayMidiDevice"))
				}
			case <-mUseMediaKeys.ClickedCh:
				if mUseMediaKeys.Checked() {
					mUseMediaKeys.Uncheck()
					viper.Set("useMediaKeys", false)
					viper.WriteConfig()
				} else {
					mUseMediaKeys.Check()
					viper.Set("useMediaKeys", true)
					viper.WriteConfig()
				}
			case <-menuExit:
				return
			}
		}
	}()

	go func() { // loop for menu item 'Quit'
		<-mQuitItem.ClickedCh
		close(quitCh)   // quit readShuttle go routine
		close(menuExit) // quit go routines for menu items
		systray.Quit()
	}()
	// Instantiate MIDI Controller
	startListeners(controlName, shuttlePro)
}

// startListeners creates and opens the specified MIDI device and starts the event handling goroutine readShuttle.
// In case the goroutine is already running it is restarted.
func startListeners(midiName string, shuttlePro *devices.ShuttleProV2) {
	if quitCh != nil {
		close(quitCh)
		mControl.close()
	}
	quitCh = make(chan struct{})

	mControl = newMIDIController(nil, midiName, messageRepeatDelay*time.Millisecond, 0)
	if err := mControl.open(); err != nil {
		dlgs.Error(applicationName, "Unable to open MIDI device. Please select the correct device in the context menu.\n"+err.Error())
	} else {
		// init: send defaults to midi device
		if viper.GetBool("useDisplay") {
			refreshDisplay(viper.GetString("displayMidiDevice"))
		}
		mControl.sendCommand(mainVolumeCC, uint8(mainVolume), false)
		mControl.sendCommand(headPhoneVolumeCC, uint8(headPhoneVolume), false)

		go readShuttle(quitCh, shuttlePro, mControl)
	}
}

// readShuttle is the goroutine used to handle all ShuttlePro events and to send out the MIDI messages.
// The routine is stopped by closing the quitch channel
func readShuttle(quitCh chan struct{}, shuttlePro *devices.ShuttleProV2, midiController midiController) {

	// doButton executes the MIDI command and changes the display text for the given button
	doButton := func(buttonNumber int, buttonCommand int) {
		switch buttonCommand {
		case on:
			midiController.sendCommand(csPro[buttonNumber].cc, CCvalueOn, false)
			if csPro[buttonNumber].msgOn != "" {
				if csPro[buttonNumber].LCDchannel != 0 {
					if viper.GetBool("useDisplay") {
						devices.DisplayLCDtext(viper.GetString("displayMidiDevice"), csPro[buttonNumber].LCDchannel, csPro[buttonNumber].LCDrow, csPro[buttonNumber].msgOn)
					}
				}
			}
		case off:
			midiController.sendCommand(csPro[buttonNumber].cc, CCvalueOff, false)
			if csPro[buttonNumber].msgOff != "" {
				if csPro[buttonNumber].LCDchannel != 0 {
					if viper.GetBool("useDisplay") {
						devices.DisplayLCDtext(viper.GetString("displayMidiDevice"), csPro[buttonNumber].LCDchannel, csPro[buttonNumber].LCDrow, csPro[buttonNumber].msgOff)
					}
				}
			}
		case toggle:
			if csPro[buttonNumber].latch {
				if csPro[buttonNumber].state {
					midiController.sendCommand(csPro[buttonNumber].cc, CCvalueOff, false)
					if csPro[buttonNumber].msgOff != "" {
						if csPro[buttonNumber].LCDchannel != 0 {
							if viper.GetBool("useDisplay") {
								devices.DisplayLCDtext(viper.GetString("displayMidiDevice"), csPro[buttonNumber].LCDchannel, csPro[buttonNumber].LCDrow, csPro[buttonNumber].msgOff)
							}
						}
					}
				} else {
					midiController.sendCommand(csPro[buttonNumber].cc, CCvalueOn, false)
					if csPro[buttonNumber].msgOn != "" {
						if csPro[buttonNumber].LCDchannel != 0 {
							if viper.GetBool("useDisplay") {
								devices.DisplayLCDtext(viper.GetString("displayMidiDevice"), csPro[buttonNumber].LCDchannel, csPro[buttonNumber].LCDrow, csPro[buttonNumber].msgOn)
							}
						}
					}
				}
				csPro[buttonNumber].state = !csPro[buttonNumber].state
			}
		default:
		}
	}

	shuttlePro.WheelPosition = make(chan int8)
	shuttlePro.DialDirection = make(chan int8)

	shuttlePro.Button1Pressed = make(chan bool)
	shuttlePro.Button2Pressed = make(chan bool)
	shuttlePro.Button3Pressed = make(chan bool)
	shuttlePro.Button4Pressed = make(chan bool)
	shuttlePro.Button5Pressed = make(chan bool)
	shuttlePro.Button6Pressed = make(chan bool)
	shuttlePro.Button7Pressed = make(chan bool)
	shuttlePro.Button8Pressed = make(chan bool)
	shuttlePro.Button9Pressed = make(chan bool)
	shuttlePro.Button10Pressed = make(chan bool)
	shuttlePro.Button11Pressed = make(chan bool)
	shuttlePro.Button12Pressed = make(chan bool)
	shuttlePro.Button13Pressed = make(chan bool)
	shuttlePro.Button14Pressed = make(chan bool)
	shuttlePro.Button15Pressed = make(chan bool)

	for {
		select {
		case <-quitCh:
			return
		case wp := <-shuttlePro.WheelPosition:
			if wp > 0 && wp <= 7 {
				// Invert positive wheel positions
				midiController.sendCommand(0, uint8(18*(8-wp)), true)
			} else if wp >= -7 && wp < 0 {
				midiController.sendCommand(1, uint8(18*(-wp)), true)
			} else {
				midiController.sendCommand(0, 255, false)
				midiController.sendCommand(1, 255, false)
			}
		case dd := <-shuttlePro.DialDirection:
			if dd == 1 { // clockwise: increase value
				if csPro[headPhoneButton].state { // headPhones on
					headPhoneVolume = headPhoneVolume + headPhoneVolumeDelta
					if headPhoneVolume > 127 {
						headPhoneVolume = 127
					}
					if viper.GetBool("useDisplay") {
						devices.DisplayLCDtext(viper.GetString("displayMidiDevice"), csPro[headPhoneButton].LCDchannel, lowerRow, headPhoneVolTable[uint8(headPhoneVolume)])
					}
					midiController.sendCommand(headPhoneVolumeCC, uint8(headPhoneVolume), false)
				} else { // headPhones off
					mainVolume = mainVolume + mainVolumeDelta
					if mainVolume > 127 {
						mainVolume = 127
					}
					if viper.GetBool("useDisplay") {
						devices.DisplayLCDtext(viper.GetString("displayMidiDevice"), csPro[LRbutton].LCDchannel, lowerRow, mainVolTable[uint8(mainVolume)])
					}
					midiController.sendCommand(mainVolumeCC, uint8(mainVolume), false)
				}
			} else { // counter clockwise: decrease value
				if csPro[headPhoneButton].state { // headPhones on
					if headPhoneVolume > 0 {
						headPhoneVolume = headPhoneVolume - headPhoneVolumeDelta
					}
					if viper.GetBool("useDisplay") {
						devices.DisplayLCDtext(viper.GetString("displayMidiDevice"), csPro[headPhoneButton].LCDchannel, lowerRow, headPhoneVolTable[uint8(headPhoneVolume)])
					}
					midiController.sendCommand(headPhoneVolumeCC, uint8(headPhoneVolume), false)
				} else {
					if mainVolume > 0 {
						mainVolume = mainVolume - mainVolumeDelta
					}
					if viper.GetBool("useDisplay") {
						devices.DisplayLCDtext(viper.GetString("displayMidiDevice"), csPro[LRbutton].LCDchannel, lowerRow, mainVolTable[uint8(mainVolume)])
					}
					midiController.sendCommand(mainVolumeCC, uint8(mainVolume), false)
				}
			}
		case b0 := <-shuttlePro.Button1Pressed: // only toggle main if headPhones are off
			if b0 {
				if !csPro[headPhoneButton].state {
					doButton(LRbutton, toggle)
				}
			}
		case b1 := <-shuttlePro.Button2Pressed:
			if b1 {
				doButton(LFEbutton, toggle)
			}
		case b2 := <-shuttlePro.Button3Pressed:
			if b2 {
				doButton(LsRsButton, toggle)
			}
		case b3 := <-shuttlePro.Button4Pressed:
			if b3 {
				if csPro[headPhoneButton].state { // headPhone -> off, LR + LFE -> on
					if csPro[LRbutton].state { // back to previous state
						doButton(LRbutton, on)
					}
					if csPro[LFEbutton].state {
						doButton(LFEbutton, on)
					}
					if csPro[LsRsButton].state {
						doButton(LsRsButton, on)
					}
				} else { // turn headPhones on, everything else off
					doButton(LRbutton, off)
					doButton(LFEbutton, off)
					doButton(LsRsButton, off)
				}
				doButton(headPhoneButton, toggle)
			}
		case b4 := <-shuttlePro.Button5Pressed: // previous
			if b4 {
				if viper.GetBool("useMediaKeys") {
					sendMediaKey(previous)
					break
				}
				doButton(4, on)
			}
		case b5 := <-shuttlePro.Button6Pressed: // next
			if b5 {
				if viper.GetBool("useMediaKeys") {
					sendMediaKey(next)
					break
				}
				doButton(5, on)
			}
		case b6 := <-shuttlePro.Button7Pressed: // stop
			if b6 {
				if viper.GetBool("useMediaKeys") {
					sendMediaKey(stop)
					break
				}
				doButton(6, on)
			}
		case b7 := <-shuttlePro.Button8Pressed: // play
			if b7 {
				if viper.GetBool("useMediaKeys") {
					sendMediaKey(play)
					break
				}
				doButton(7, on)
			}
		case b8 := <-shuttlePro.Button9Pressed:
			if b8 {
				if csPro[stereoSurroundButton].state { // surroundMode -> off, LsRs to previous state
					if !csPro[LsRsButton].state {
						doButton(LsRsButton, off)
					}
				} else { // surroundMode -> on, turn LsRs on
					doButton(LsRsButton, on)
				}
				doButton(stereoSurroundButton, toggle)
			}
		case b9 := <-shuttlePro.Button10Pressed:
			if b9 {
				doButton(9, toggle)
			}
		case b10 := <-shuttlePro.Button11Pressed:
			if b10 {
				doButton(10, toggle)
			}
		case b11 := <-shuttlePro.Button12Pressed:
			if b11 {
				doButton(11, toggle)
			}
		case b12 := <-shuttlePro.Button13Pressed:
			if b12 {
				doButton(12, toggle)
			}
		case b13 := <-shuttlePro.Button14Pressed:
			if b13 {
				doButton(13, toggle)
			}
		case b14 := <-shuttlePro.Button15Pressed:
			if b14 {
				doButton(14, toggle)
			}
		}
	}
}

// onExit is called by systray on exit and closes the MidiController
func onExit() {
	if mControl != nil {
		mControl.close()
	}
}

func main() {
	if err := initSettings(); err != nil {
		slog.Error("initSettings not successful: ", err)
		return
	}
	systray.Run(onReady, onExit)
}
