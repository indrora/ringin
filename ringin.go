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
}

func loadConfig(path string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	return &config, err
}

func main() {
	writeDefaults := flag.Bool("write-defaults", false, "Write default configuration to specified config file")
	configPath := flag.String("config", "ringin.toml", "Path to configuration file")
	flag.Parse()

	if *writeDefaults {
		if err := writeDefaultConfig(*configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write default config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Default configuration written to %s\n", *configPath)
		os.Exit(0)
	}
	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	mode := &serial.Mode{
		BaudRate: config.Serial.BaudRate,
		DataBits: config.Serial.DataBits,
	}

	// Set parity based on config
	switch strings.ToUpper(config.Serial.Parity) {
	case "N":
		mode.Parity = serial.NoParity
	case "E":
		mode.Parity = serial.EvenParity
	case "O":
		mode.Parity = serial.OddParity
	}

	// Set stop bits based on config
	switch config.Serial.StopBits {
	case 1:
		mode.StopBits = serial.OneStopBit
	case 2:
		mode.StopBits = serial.TwoStopBits
	}

	port, err := serial.Open(config.Serial.Port, mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open serial port: %v\n", err)
		os.Exit(1)
	}
	defer port.Close()

	// Initialize modem using commands from config
	for _, cmd := range config.Modem.InitCommands {
		_, err := port.Write([]byte(cmd + "\r"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send command %s: %v\n", cmd, err)
			os.Exit(1)
		}
		time.Sleep(500 * time.Millisecond)
	}

	reader := bufio.NewReader(port)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from port: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)

		switch {
		case strings.Contains(line, "RING"):
			fmt.Println("Incoming call detected")

		case strings.Contains(line, "CONNECT"):
			fmt.Println("Connection established")

			cmd := exec.Command(config.Program.Command, config.Program.Args...)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create stdin pipe: %v\n", err)
				continue
			}

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create stdout pipe: %v\n", err)
				continue
			}

			if err := cmd.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to start subprocess: %v\n", err)
				continue
			}

			go func() {
				io.Copy(port, stdout)
			}()

			go func() {
				io.Copy(stdin, port)
			}()

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
			port.Write([]byte("ATH0\r"))
			time.Sleep(500 * time.Millisecond)

		case strings.Contains(line, "NO CARRIER"):
			fmt.Println("Connection terminated")
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
