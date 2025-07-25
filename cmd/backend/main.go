package main

import (
	"errors"
	"fmt"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/config"
	"go.uber.org/zap"
	"log"
)

func main() {
	// initialize configuration
	cfg, cfgLog := config.Load()
	err := cfg.Validate()
	if err != nil {
		if errors.Is(err, config.ErrDatabaseURLRequired) {
			title := "Databae URL is required"
			message := "Please set the DATABASE_URL environment variable or provide it in the configuration file."
			message = EarlyApplicationFailed(title, message)
			log.Fatal(message)
		} else {
			log.Fatalf("failed to validate configuration: %v", err)
		}
	}

	// initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	cfgLog.FlushToZap(logger)

	logger.Info("Logger initialized successfully")

	logger.Info("Successfully shutdown")
}

func EarlyApplicationFailed(title, action string) string {
	result := `
-----------------------------------------
Application Failed to Start
-----------------------------------------

# What's wrong?
%s

# How to fix it?
%s

`

	result = fmt.Sprintf(result, title, action)
	return result
}
