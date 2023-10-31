package main

const (
	applicationName = "ShuttleMidi"
	// midi CC number for main volume
	mainVolumeCC = 7
	// midi CC number for headPhone volume
	headPhoneVolumeCC = 102

	// button number from ShuttlePro device
	LRbutton             = 0
	LFEbutton            = 1
	LsRsButton           = 2
	headPhoneButton      = 3
	stereoSurroundButton = 8

	CCvalueOn  = 127
	CCvalueOff = 0

	off    = 0
	on     = 1
	toggle = 2

	upperRow = 0
	lowerRow = 56 // offset for lower LCD row

	// delay in milliseconds for repeating midi commands
	messageRepeatDelay = 300
	// delay in milliseconds between sending upper & lower row of text to MCU device
	displayRowsDelay = 400
)

var (
	// values of dB levels from RME TotalMix master fader per MIDI CC 7 step
	mainVolTable [128]string = [128]string{
		"  -oo  ", " -63.2 ", " -62.0 ", " -60.9 ", " -59.7 ", " -58.6 ", " -57.5 ", " -56.3 ", " -55.2 ", " -54.1 ", // 0
		" -53.1 ", " -52.0 ", " -50.9 ", " -49.9 ", " -48.9 ", " -47.8 ", " -46.8 ", " -45.8 ", " -44.8 ", " -43.8 ", // 10
		" -42.9 ", " -41.9 ", " -41.0 ", " -40.1 ", " -39.1 ", " -38.2 ", " -37.3 ", " -36.4 ", " -35.6 ", " -34.7 ", // 20
		" -33.9 ", " -33.0 ", " -32.2 ", " -31.4 ", " -30.6 ", " -29.8 ", " -29.0 ", " -28.2 ", " -27.5 ", " -26.7 ", // 30
		" -26.0 ", " -25.3 ", " -24.6 ", " -23.9 ", " -23.2 ", " -22.5 ", " -21.8 ", " -21.2 ", " -20.5 ", " -19.9 ", // 40
		" -19.3 ", " -18.7 ", " -18.1 ", " -17.5 ", " -16.9 ", " -16.4 ", " -15.8 ", " -15.3 ", " -14.8 ", " -14.3 ", // 50
		" -13.8 ", " -13.3 ", " -12.8 ", " -12.3 ", " -11.9 ", " -11.4 ", " -11.0 ", " -10.6 ", " -10.2 ", "  -9.8 ", // 60
		"  -9.4 ", "  -9.0 ", "  -8.6 ", "  -8.3 ", "  -8.0 ", "  -7.6 ", "  -7.3 ", "  -7.0 ", "  -6.7 ", "  -6.4 ", // 70
		"  -6.2 ", "  -5.9 ", "  -5.6 ", "  -5.4 ", "  -5.1 ", "  -4.9 ", "  -4.6 ", "  -4.4 ", "  -4.1 ", "  -3.9 ", // 80
		"  -3.6 ", "  -3.3 ", "  -3.1 ", "  -2.8 ", "  -2.6 ", "  -2.3 ", "  -2.1 ", "  -1.8 ", "  -1.5 ", "  -1.3 ", // 90
		"  -1.0 ", "  -0.8 ", "  -0.5 ", "  -0.3 ", "   0.0 ", "   0.3 ", "   0.5 ", "   0.8 ", "   1.0 ", "   1.3 ", // 100
		"   1.5 ", "   1.8 ", "   2.1 ", "   2.3 ", "   2.6 ", "   2.8 ", "   3.1 ", "   3.3 ", "   3.6 ", "   3.9 ", // 110
		"   4.1 ", "   4.4 ", "   4.6 ", "   4.9 ", "   5.1 ", "   5.4 ", "   5.6 ", "  6.0  ", // 120
	}

	// volume value in percentage
	headPhoneVolTable [128]string = [128]string{
		"  0.00 ", "  0.79 ", "  1.57 ", "  2.36 ", "  3.15 ", "  3.94 ", "  4.72 ", "  5.51 ", "  6.30 ", "  7.09 ",
		"  7.87 ", "  8.66 ", "  9.45 ", " 10.24 ", " 11.02 ", " 11.81 ", " 12.60 ", " 13.39 ", " 14.17 ", " 14.96 ",
		" 15.75 ", " 16.54 ", " 17.32 ", " 18.11 ", " 18.90 ", " 19.69 ", " 20.47 ", " 21.26 ", " 22.05 ", " 22.83 ",
		" 23.62 ", " 24.41 ", " 25.20 ", " 25.98 ", " 26.77 ", " 27.56 ", " 28.35 ", " 29.13 ", " 29.92 ", " 30.71 ",
		" 31.50 ", " 32.28 ", " 33.07 ", " 33.86 ", " 34.65 ", " 35.43 ", " 36.22 ", " 37.01 ", " 37.80 ", " 38.58 ",
		" 39.37 ", " 40.16 ", " 40.94 ", " 41.73 ", " 42.52 ", " 43.31 ", " 44.09 ", " 44.88 ", " 45.67 ", " 46.46 ",
		" 47.24 ", " 48.03 ", " 48.82 ", " 49.61 ", " 50.39 ", " 51.18 ", " 51.97 ", " 52.76 ", " 53.54 ", " 54.33 ",
		" 55.12 ", " 55.91 ", " 56.69 ", " 57.48 ", " 58.27 ", " 59.06 ", " 59.84 ", " 60.63 ", " 61.42 ", " 62.20 ",
		" 62.99 ", " 63.78 ", " 64.57 ", " 65.35 ", " 66.14 ", " 66.93 ", " 67.72 ", " 68.50 ", " 69.29 ", " 70.08 ",
		" 70.87 ", " 71.65 ", " 72.44 ", " 73.23 ", " 74.02 ", " 74.80 ", " 75.59 ", " 76.38 ", " 77.17 ", " 77.95 ",
		" 78.74 ", " 79.53 ", " 80.31 ", " 81.10 ", " 81.89 ", " 82.68 ", " 83.46 ", " 84.25 ", " 85.04 ", " 85.83 ",
		" 86.61 ", " 87.40 ", " 88.19 ", " 88.98 ", " 89.76 ", " 90.55 ", " 91.34 ", " 92.13 ", " 92.91 ", " 93.70 ",
		" 94.49 ", " 95.28 ", " 96.06 ", " 96.85 ", " 97.64 ", " 98.43 ", " 99.21 ", "100.00 ",
	}

	// start value for main volume
	mainVolume float32 = 40
	// amount of main volume change per dial step
	mainVolumeDelta float32 = 1.3
	// start value for headPhone volume
	headPhoneVolume float32 = 60
	// amount of headphone volume change per dial step
	headPhoneVolumeDelta float32 = 1.4
	// configDefaults contains the default configuration written to the configuration file
	configDefaults = map[string]interface{}{
		"controlMidiDevice": "IAC monitorControl",
		"displayMidiDevice": "X-Touch INT",
		"useDisplay":        true,
		"useMediaKeys":      true,
	}
)

type button struct {
	state      bool   // (initial) state of the button
	cc         uint8  // midi CC number send for this button
	latch      bool   // does this button toggle its state?
	msgOn      string // GUI message for 'on'
	msgOff     string // GUI message for 'off'
	LCDchannel uint8  // channel to display text on MCU
	LCDrow     uint8  // upper or lower row on LCD
}

var csPro [15]button = [15]button{

	{ // 00 LR
		state:      true,
		cc:         70,
		latch:      true,
		msgOn:      " LR on ",
		msgOff:     " LR off",
		LCDchannel: 5,
		LCDrow:     upperRow,
	},
	{ // 01 LFE
		state:      true,
		cc:         71,
		latch:      true,
		msgOn:      "LFE on ",
		msgOff:     "LFE off",
		LCDchannel: 6,
		LCDrow:     lowerRow,
	},
	{ // 02 LsRs
		state:      true,
		cc:         72,
		latch:      true,
		msgOn:      "LRs on ",
		msgOff:     "LRs off",
		LCDchannel: 6,
		LCDrow:     upperRow,
	},
	{ // 03 HeadPhones
		state:      false,
		cc:         73,
		latch:      true,
		msgOn:      "Phn on ",
		msgOff:     "Phn off",
		LCDchannel: 7,
		LCDrow:     upperRow,
	},
	{ // 04 Previous
		state:      true,
		cc:         74,
		latch:      false,
		msgOn:      "",
		msgOff:     "",
		LCDchannel: 0,
	},
	{ // 05 Next
		state:      true,
		cc:         75,
		latch:      false,
		msgOn:      "",
		msgOff:     "",
		LCDchannel: 0,
	},
	{ // 06 Stop
		state:      true,
		cc:         76,
		latch:      false,
		msgOn:      "",
		msgOff:     "",
		LCDchannel: 0,
	},
	{ // 07 Play
		state:      true,
		cc:         77,
		latch:      false,
		msgOn:      "",
		msgOff:     "",
		LCDchannel: 0,
	},
	{ // 08 Record / stereoSurround
		state:      false,
		cc:         78,
		latch:      true,
		msgOn:      "Surrnd ",
		msgOff:     "Stereo ",
		LCDchannel: 8,
		LCDrow:     upperRow,
	},
	{ // 09 Left solo
		state:      false,
		cc:         79,
		latch:      true,
		msgOn:      " -Left-",
		msgOff:     "       ",
		LCDchannel: 8,
		LCDrow:     lowerRow,
	},
	{ // 10 Right solo
		state:      false,
		cc:         80,
		latch:      true,
		msgOn:      "-Right-",
		msgOff:     "       ",
		LCDchannel: 8,
		LCDrow:     lowerRow,
	},
	{ // 11 Mid solo
		state:      false,
		cc:         81,
		latch:      true,
		msgOn:      " -Mid- ",
		msgOff:     "       ",
		LCDchannel: 8,
		LCDrow:     lowerRow,
	},
	{ // 12 Side solo
		state:      false,
		cc:         82,
		latch:      true,
		msgOn:      " -Side-",
		msgOff:     "       ",
		LCDchannel: 8,
		LCDrow:     lowerRow,
	},
	{ // 13 Dim
		state:      false,
		cc:         83,
		latch:      true,
		msgOn:      " -Dim- ",
		msgOff:     "       ",
		LCDchannel: 8,
		LCDrow:     lowerRow,
	},
	{ // 14 Mono
		state:      false,
		cc:         84,
		latch:      true,
		msgOn:      " -Mono-",
		msgOff:     "       ",
		LCDchannel: 8,
		LCDrow:     lowerRow,
	},
}
