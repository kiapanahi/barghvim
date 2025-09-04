package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MoKhajavi75/barghvim/internal/server"
	"github.com/MoKhajavi75/barghvim/pkg/logging"
	"github.com/MoKhajavi75/barghvim/pkg/telemetry"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	
	// Set up structured logging
	if os.Getenv("LOG_LEVEL") == "debug" {
		logging.SetLevel(logrus.DebugLevel)
	}
	
	// Initialize telemetry
	telConfig := telemetry.DefaultConfig()
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		telConfig.Environment = env
	}
	if endpoint := os.Getenv("OTLP_ENDPOINT"); endpoint != "" {
		telConfig.OTLPEndpoint = endpoint
	}
	
	tel, err := telemetry.Setup(ctx, telConfig)
	if err != nil {
		log.Fatal("Failed to set up telemetry:", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tel.Shutdown(shutdownCtx); err != nil {
			logging.Error(shutdownCtx, "Failed to shutdown telemetry", err)
		}
	}()
	
	// Start the server
	r := server.New(tel)
	
	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	
	// Start server in a goroutine
	go func() {
		logging.LogServiceStart(ctx, telConfig.ServiceName, telConfig.ServiceVersion, 8080)
		if err := r.Run(":8080"); err != nil {
			logging.Fatal(ctx, "Server failed to start", err)
		}
	}()
	
	// Wait for interrupt signal
	<-stop
	logging.LogServiceStop(ctx, telConfig.ServiceName)
}
