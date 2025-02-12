package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"go.bug.st/serial"
)

type Config struct {
	Serial struct {
		Port     string `toml:"port"`
		BaudRate int    `toml:"baud_rate"`
		DataBits int    `toml:"data_bits"`
		Parity   string `toml:"parity"`
		StopBits int    `toml:"stop_bits"`
	} `toml:"serial"`

	Modem struct {
		InitCommands []string `toml:"init_commands"`
	} `toml:"modem"`

	Program struct {
		Command string   `toml:"command"`
		Args    []string `toml:"args"`
	} `toml:"program"`

	mode *serial.Mode // Non-exported field for parsed mode
}

func loadConfig(path string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}

	// Parse parity
	var mode serial.Mode
	config.mode = &mode

	switch strings.ToUpper(config.Serial.Parity) {
	case "N":
		mode.Parity = serial.NoParity
	case "E":
		mode.Parity = serial.EvenParity
	case "O":
		mode.Parity = serial.OddParity
	default:
		return nil, fmt.Errorf("invalid parity setting")
	}

	// Parse stop bits
	switch config.Serial.StopBits {
	case 1:
		mode.StopBits = serial.OneStopBit
	case 2:
		mode.StopBits = serial.TwoStopBits
	default:
		return nil, fmt.Errorf("invalid stop bits setting")
	}
	// Set other mode fields
	config.mode.BaudRate = config.Serial.BaudRate
	config.mode.DataBits = config.Serial.DataBits

	return &config, nil
}

func writeDefaultConfig(path string) error {
	defaultPort := "/dev/ttyUSB0"
	if runtime.GOOS == "windows" {
		defaultPort = "COM1"
	}

	config := Config{}
	config.Serial.Port = defaultPort
	config.Serial.BaudRate = 9600
	config.Serial.DataBits = 8
	config.Serial.Parity = "N"
	config.Serial.StopBits = 1

	config.Modem.InitCommands = []string{
		"ATZ",    // Reset modem
		"ATH0",   // Hang up if connected
		"ATS0=1", // Auto-answer after 1 ring
		"AT&K0",  // Disable flow control
	}

	config.Program.Command = "fortune"
	config.Program.Args = []string{}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(config)
}
