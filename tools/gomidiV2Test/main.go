package main

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

func main() {
	defer midi.CloseDriver()

	out, err := midi.FindOutPort("X-Touch INT")
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	displayLCDtext(&out, 4, 0, "Hey!r  ")
	//fmt.Println(out.String())

}

func displayLCDtext(out *drivers.Out, channel uint8, row uint8, text string) {
	send, err := midi.SendTo(*out)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	msg := []byte{}

	startSysEx := uint8(240)
	vendor := []byte{0, 0, 102, 20, 18}
	start := uint8(row + ((channel - 1) * 7))
	textByte := []byte(text)
	endSysEx := uint8(247)

	msg = append(msg, startSysEx)
	msg = append(msg, vendor...)
	msg = append(msg, start)
	msg = append(msg, textByte...)
	msg = append(msg, endSysEx)

	send(msg)
}
