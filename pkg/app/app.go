package app

import (
	"context"
	"net/http"

	"github.com/MoKhajavi75/barghvim/internal/server"
	"github.com/MoKhajavi75/barghvim/pkg/telemetry"
)

func Handler() http.Handler {
	// For the Vercel deployment, use a minimal telemetry setup
	tel, err := telemetry.Setup(context.Background(), telemetry.TelemetryConfig{
		ServiceName:    "barghvim",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		OTLPEndpoint:   "", // No OTLP for Vercel
		EnableTracing:  false, // Minimal overhead for serverless
		EnableMetrics:  false, // Minimal overhead for serverless
	})
	
	if err != nil {
		// Fall back to mock telemetry if setup fails
		tel = telemetry.NewMockTelemetry()
	}
	
	return server.New(tel)
}
