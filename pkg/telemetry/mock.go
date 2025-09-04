package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

// NewMockTelemetry creates a mock telemetry instance for testing
func NewMockTelemetry() *Telemetry {
	return &Telemetry{
		TracerProvider: trace.NewNoopTracerProvider(),
		MeterProvider:  noop.NewMeterProvider(),
		Tracer:         otel.Tracer("test"),
		Meter:          noop.NewMeterProvider().Meter("test"),
		shutdownFuncs:  []func(context.Context) error{},
	}
}
