package main

import (
	"flag"
	"fmt"

	"log"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewDevelopment()
	defer logger.Sync()
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
		return
	}

	config, err := loadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
		return
	}

	if err := runModemLoop(config); err != nil {
		log.Fatalf("Modem loop failed", zap.Error(err))
	}
}
