package devices

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

const (
	// Behringer X-Touch
	MCUvendorId      = "00 00 66 14 12"
	displayRowsDelay = 400
)

// TODO: replace sendmidi cmd with gomidi

func DisplayLCDtext(channel uint8, row uint8, text string) {
	if !viper.GetBool("useDisplay") {
		return
	}

	hexText := ""
	for _, v := range text {
		hexText = hexText + fmt.Sprintf(" %X", v)
	}

	start := row + ((channel - 1) * 7)
	app := "/opt/homebrew/bin/sendmidi dev '" + viper.GetString("displayMidiDevice") + "' hex syx " + MCUvendorId + fmt.Sprintf(" %X", start) + hexText
	cmd := exec.Command("bash", "-c", app)
	if err := cmd.Run(); err != nil {
		slog.Error("can't run 'sendmidi'", err)
	}
}

func ClearDisplay(channel uint8) {
	hexText := ""
	text := "                            "
	for _, v := range text {
		hexText = hexText + fmt.Sprintf(" %X", v)
	}

	start := 0 + ((channel - 1) * 7) // upper row
	app := "/opt/homebrew/bin/sendmidi dev '" + viper.GetString("displayMidiDevice") + "' hex syx " + MCUvendorId + fmt.Sprintf(" %X", start) + hexText
	cmd := exec.Command("bash", "-c", app)
	if err := cmd.Run(); err != nil {
		slog.Error("can't run 'sendmidi'", err)
	}

	time.Sleep(displayRowsDelay * time.Millisecond) // wait a little bit, MCU device might be 'overwhelmed'

	start = 56 + ((channel - 1) * 7) // lower row
	app = "/opt/homebrew/bin/sendmidi dev '" + viper.GetString("displayMidiDevice") + "' hex syx " + MCUvendorId + fmt.Sprintf(" %X", start) + hexText
	cmd = exec.Command("bash", "-c", app)
	if err := cmd.Run(); err != nil {
		slog.Error("can't run 'sendmidi'", err)
	}
}
