package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.uber.org/zap"
)

func runModemLoop(config *Config) error {
	port, err := serial.Open(config.Serial.Port, config.mode)
	if err != nil {
		return fmt.Errorf("error opening port: %w", err)
	}
	defer port.Close()

	reader := bufio.NewReader(port)

	time.Sleep(500 * time.Millisecond)
	for {

		// Send init commands
		logger.Info("Sending init commands")

		for _, cmd := range config.Modem.InitCommands {
			logger.Debug("Sending command", zap.String("command", cmd))
			port.Write([]byte(cmd + "\r\n"))
			time.Sleep(50 * time.Millisecond)
		}

		port.Write([]byte("\n"))

		for reader.Buffered() < 2 {
			time.Sleep(1 * time.Second)
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading from port: %w", err)
		}

		logger.Debug("Incoming line from modem", zap.String("line", line))

		switch strings.TrimSpace(line) {
		case "ATH0":
		case "\r\n":
			port.Drain()
			// ... Don't worry about it
		case "OK":
			logger.Info("Modem liked our command")
			time.Sleep(100 * time.Millisecond)
			port.Drain()
		case "NO CARRIER":
			fmt.Println("Connection terminated")
			continue
		case "RING":
			logger.Info("Incoming call...")
			port.Write([]byte("ATA\r\n"))
			logger.Debug("Sent ATA")
			port.Drain()
		case "ERROR":
			logger.Error("Modem didn't like our command....")
			time.Sleep(100 * time.Millisecond)
		case "CONNECT":
			logger.Info("Line is connected, welcome to zombocom")
			handleRing(&port, reader, config.Program.Command, config.Program.Args)

			//  Wake up the modem
			(port).Write([]byte("+++\r"))
			// Wait for the modem to catch up with us.
			time.Sleep(1500 * time.Millisecond)
			bb, err := reader.ReadString('\n')
			if err != nil {
				logger.Error("Error reading line from modem", zap.Error(err))
			}
			if bb == "OK" {
				logger.Info("Modem is ready")
			} else {
				logger.Info("Modem angy")
				panic("Modem is angy")

			}
			(port).Write([]byte("ATH0\r"))
			time.Sleep(1500 * time.Millisecond)

		default:

			logger.Info("Received line from modem", zap.String("line", line))
			time.Sleep(100 * time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond)
		port.Drain()
	}
}
