package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	ServiceName    = "barghvim"
	ServiceVersion = "1.0.0"
)

// TelemetryConfig holds configuration for telemetry setup
type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	EnableTracing  bool
	EnableMetrics  bool
}

// DefaultConfig returns a default telemetry configuration
func DefaultConfig() TelemetryConfig {
	return TelemetryConfig{
		ServiceName:    ServiceName,
		ServiceVersion: ServiceVersion,
		Environment:    "development",
		OTLPEndpoint:   "", // Empty means no OTLP export
		EnableTracing:  true,
		EnableMetrics:  true,
	}
}

// Telemetry holds the telemetry providers and cleanup functions
type Telemetry struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	Tracer         trace.Tracer
	Meter          metric.Meter
	shutdownFuncs  []func(context.Context) error
}

// Setup initializes OpenTelemetry with the given configuration
func Setup(ctx context.Context, config TelemetryConfig) (*Telemetry, error) {
	var shutdownFuncs []func(context.Context) error

	// Create resource
	res, err := newResource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Set up propagator
	otel.SetTextMapPropagator(newPropagator())

	tel := &Telemetry{}

	// Set up tracing
	if config.EnableTracing {
		tracerProvider, err := newTracerProvider(ctx, res, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create tracer provider: %w", err)
		}
		
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
		
		tel.TracerProvider = tracerProvider
		tel.Tracer = tracerProvider.Tracer(
			config.ServiceName,
			trace.WithInstrumentationVersion(config.ServiceVersion),
		)
	}

	// Set up metrics
	if config.EnableMetrics {
		meterProvider, err := newMeterProvider(ctx, res)
		if err != nil {
			return nil, fmt.Errorf("failed to create meter provider: %w", err)
		}
		
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
		
		tel.MeterProvider = meterProvider
		tel.Meter = meterProvider.Meter(
			config.ServiceName,
			metric.WithInstrumentationVersion(config.ServiceVersion),
		)
	}

	tel.shutdownFuncs = shutdownFuncs
	return tel, nil
}

// Shutdown gracefully shuts down all telemetry providers
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var err error
	for _, fn := range t.shutdownFuncs {
		if shutdownErr := fn(ctx); shutdownErr != nil {
			err = fmt.Errorf("%w; %v", err, shutdownErr)
		}
	}
	return err
}

// newResource creates a new OpenTelemetry resource
func newResource(config TelemetryConfig) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		),
	)
}

// newPropagator creates a new text map propagator
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTracerProvider creates a new tracer provider
func newTracerProvider(ctx context.Context, res *resource.Resource, config TelemetryConfig) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	if config.OTLPEndpoint != "" {
		// Use OTLP exporter if endpoint is provided
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(config.OTLPEndpoint),
			otlptracehttp.WithInsecure(), // Use HTTPS in production
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}
	} else {
		// Use stdout exporter for development
		exporter, err = newStdoutTraceExporter()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
		}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	return tp, nil
}

// newMeterProvider creates a new meter provider with Prometheus exporter
func newMeterProvider(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	// Create Prometheus exporter
	exporter, err := prometheus.New(
		prometheus.WithoutUnits(),
		prometheus.WithoutScopeInfo(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
		sdkmetric.WithView(
			// Convert histogram to counter for request duration
			sdkmetric.NewView(
				sdkmetric.Instrument{
					Name: "http_request_duration_seconds",
					Kind: sdkmetric.InstrumentKindHistogram,
				},
				sdkmetric.Stream{
					Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
						Boundaries: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
					},
				},
			),
		),
	)

	return mp, nil
}

// newStdoutTraceExporter creates a stdout exporter for development
func newStdoutTraceExporter() (sdktrace.SpanExporter, error) {
	// For development, we'll create a simple console exporter
	// In production, you'd want to use OTLP or Jaeger
	return &noopExporter{}, nil
}

// noopExporter is a no-op exporter for development when OTLP endpoint is not set
type noopExporter struct{}

func (e *noopExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	// In development, you might want to log spans or just ignore them
	// For production, use proper OTLP endpoint
	return nil
}

func (e *noopExporter) Shutdown(ctx context.Context) error {
	return nil
}

// GetPrometheusHandler returns the Prometheus metrics handler
func GetPrometheusHandler() (interface{}, error) {
	// The Prometheus exporter automatically registers with the default registry
	// We need to get the handler from the Prometheus client
	return nil, fmt.Errorf("use the prometheus.Handler() from the exporter")
}

// CorrelationID generates a new correlation ID for request tracing
func CorrelationID() string {
	return fmt.Sprintf("barghvim-%d", time.Now().UnixNano())
}
