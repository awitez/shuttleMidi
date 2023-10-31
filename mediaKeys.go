package main

import (
	"log/slog"
	"os/exec"
)

const (

	// MediaKeys
	previous = 0
	next     = 1
	stop     = 2
	play     = 3
)

// sendMediaKey sends a play, pause or skip command to the Apple Music.app
func sendMediaKey(command int) {
	cmdString := "osascript -e 'tell application \"Music\" to "

	switch command {
	case previous:
		cmdString = cmdString + "set player position to (player position -5)'"
	case next:
		cmdString = cmdString + "set player position to (player position +5)'"
	case stop:
		cmdString = cmdString + "pause'"
	case play:
		cmdString = cmdString + "playpause'"
	default:
		slog.Info("unknown command")
		return
	}
	cmd := exec.Command("bash", "-c", cmdString)
	if err := cmd.Run(); err != nil {
		slog.Error("can't run 'osascript' ", err)
	}
}
