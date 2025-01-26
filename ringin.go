package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

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

func main() {
	writeDefaults := flag.Bool("write-defaults", false, "Write default configuration")
	configPath := flag.String("config", "ringin.toml", "Configuration file path")
	flag.Parse()

	if *writeDefaults {
		err := writeDefaultConfig(*configPath)
		if err != nil {
			fmt.Println("Error writing default config:", err)
			return
		}
	}

	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	if err := runModemLoop(config); err != nil {
		fmt.Println("Error in modem loop:", err)
	}
}

func handleRing(port *serial.Port, reader *bufio.Reader, command string, args []string) {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error creating stdin pipe:", err)
		return
	}

	go func() {
		io.Copy(stdin, reader)
		stdin.Close()
	}()

	cmd.Stdout = *port
	cmd.Stdin = *port
	cmd.Stderr = os.Stderr

	done := make(chan bool)
	go func() {
		cmd.Wait()
		done <- true
	}()

	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if strings.Contains(strings.TrimSpace(line), "NO CARRIER") {
				cmd.Process.Kill()
				done <- true
				return
			}
		}
	}()

	<-done
}

func runModemLoop(config *Config) error {
	port, err := serial.Open(config.Serial.Port, config.mode)
	if err != nil {
		return fmt.Errorf("error opening port: %w", err)
	}
	defer port.Close()

	reader := bufio.NewReader(port)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading from port: %w", err)
		}

		switch strings.TrimSpace(line) {
		case "ATH0":
			port.Write([]byte("ATH0\r"))
			time.Sleep(500 * time.Millisecond)

		case "NO CARRIER":
			fmt.Println("Connection terminated")

		default:
			if strings.Contains(strings.TrimSpace(line), "RING") {
				handleRing(&port, reader, config.Program.Command, config.Program.Args)
			}
		}
	}
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
