package devices

import (
	"fmt"
	"log/slog"
	"os/exec"
	"time"
)

const (
	// Behringer X-Touch sysex vendor ID
	MCUvendorId = "00 00 66 14 "
	cmdText     = "12 " // 'the following bytes are text'
	cmdColor    = "72 " // 'the following bytes are colors'

	displayRowsDelay = 400                          // in milliseconds
	sendMidi         = "/opt/homebrew/bin/sendmidi" // TODO: replace sendmidi cmd with goMidi V2
)

func InitColor(device string, color string) {
	hexText := "00 00 00 00 " + color + " " + color + " " + color + " " + color

	app := sendMidi + " dev '" + device + "' hex syx " + MCUvendorId + cmdColor + hexText
	cmd := exec.Command("bash", "-c", app)
	if err := cmd.Run(); err != nil {
		slog.Error("can't run 'sendmidi'", err)
	}
}

func DisplayLCDtext(device string, channel uint8, row uint8, text string) {
	hexText := ""
	for _, v := range text {
		hexText = hexText + fmt.Sprintf(" %X", v)
	}

	start := row + ((channel - 1) * 7)
	app := sendMidi + " dev '" + device + "' hex syx " + MCUvendorId + cmdText + fmt.Sprintf(" %X", start) + hexText
	cmd := exec.Command("bash", "-c", app)
	if err := cmd.Run(); err != nil {
		slog.Error("can't run 'sendmidi'", err)
	}
}

func ClearDisplay(device string, channel uint8) {
	hexText := ""
	text := "                            "
	for _, v := range text {
		hexText = hexText + fmt.Sprintf(" %X", v)
	}

	start := 0 + ((channel - 1) * 7) // upper row
	app := sendMidi + " dev '" + device + "' hex syx " + MCUvendorId + cmdText + fmt.Sprintf(" %X", start) + hexText
	cmd := exec.Command("bash", "-c", app)
	if err := cmd.Run(); err != nil {
		slog.Error("can't run 'sendmidi'", err)
	}

	time.Sleep(displayRowsDelay * time.Millisecond) // wait a little bit, MCU device might be 'overwhelmed'

	start = 56 + ((channel - 1) * 7) // lower row
	app = sendMidi + " dev '" + device + "' hex syx " + MCUvendorId + cmdText + fmt.Sprintf(" %X", start) + hexText
	cmd = exec.Command("bash", "-c", app)
	if err := cmd.Run(); err != nil {
		slog.Error("can't run 'sendmidi'", err)
	}
	// set LCD color to white
	hexText = "07 07 07 07 07 07 07 07"

	app = sendMidi + " dev '" + device + "' hex syx " + MCUvendorId + cmdColor + hexText
	cmd = exec.Command("bash", "-c", app)
	if err := cmd.Run(); err != nil {
		slog.Error("can't run 'sendmidi'", err)
	}
}
