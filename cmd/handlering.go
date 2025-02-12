package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"time"

	"go.bug.st/serial"
	"go.uber.org/zap"
)

func handleRing(port *serial.Port, reader *bufio.Reader, command string, args []string) {
	logger.Info("Incoming call")

	cmd := exec.Command(command, args...)

	cmd.Env = []string{"MODEM=dumb"}

	buffs := bufio.NewReadWriter(
		bufio.NewReader(*port),
		bufio.NewWriter(*port),
	)
	cmd.Stdout = io.MultiWriter(*port, os.Stdout)
	cmd.Stdin = buffs
	cmd.Stderr = os.Stderr

	done := make(chan bool)
	go func() {
		logger.Info("Executing command", zap.String("command", command))
		cmd.Start()
		cmd.Wait()
		logger.Info("Command execution completed")
		done <- true
	}()

	go func() {
		for {
			modembits, err := (*port).GetModemStatusBits()
			if err != nil {
				logger.Error("Error getting modem status bits", zap.Error(err))
			}

			if !modembits.DCD {
				logger.Info("DCD is off, we're done here")
				_ = cmd.Process.Kill()
				done <- true
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	<-done

}
