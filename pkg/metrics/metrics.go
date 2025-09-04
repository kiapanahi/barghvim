package metrics

import (
	"context"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds all the metric instruments
type Metrics struct {
	HTTPRequestsTotal    metric.Int64Counter
	HTTPRequestDuration  metric.Float64Histogram
	HTTPRequestsInFlight metric.Int64UpDownCounter

	APICallsTotal    metric.Int64Counter
	APICallDuration  metric.Float64Histogram
	APICallsFailures metric.Int64Counter

	CalendarGenerated metric.Int64Counter
	OutagesFetched    metric.Int64Counter

	ProcessStartTime metric.Float64Gauge
}

// New creates and registers all metrics
func New(meter metric.Meter) (*Metrics, error) {
	m := &Metrics{}

	var err error

	// HTTP metrics
	m.HTTPRequestsTotal, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	m.HTTPRequestDuration, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.HTTPRequestsInFlight, err = meter.Int64UpDownCounter(
		"http_requests_in_flight",
		metric.WithDescription("Number of HTTP requests currently being processed"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	// API call metrics
	m.APICallsTotal, err = meter.Int64Counter(
		"api_calls_total",
		metric.WithDescription("Total number of external API calls"),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return nil, err
	}

	m.APICallDuration, err = meter.Float64Histogram(
		"api_call_duration_seconds",
		metric.WithDescription("Duration of external API calls in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.APICallsFailures, err = meter.Int64Counter(
		"api_calls_failures_total",
		metric.WithDescription("Total number of failed external API calls"),
		metric.WithUnit("{failure}"),
	)
	if err != nil {
		return nil, err
	}

	// Business logic metrics
	m.CalendarGenerated, err = meter.Int64Counter(
		"calendars_generated_total",
		metric.WithDescription("Total number of calendars generated"),
		metric.WithUnit("{calendar}"),
	)
	if err != nil {
		return nil, err
	}

	m.OutagesFetched, err = meter.Int64Counter(
		"outages_fetched_total",
		metric.WithDescription("Total number of outages fetched"),
		metric.WithUnit("{outage}"),
	)
	if err != nil {
		return nil, err
	}

	// Process metrics
	m.ProcessStartTime, err = meter.Float64Gauge(
		"process_start_time_seconds",
		metric.WithDescription("Start time of the process since Unix epoch in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// Record process start time
	m.ProcessStartTime.Record(context.Background(), float64(time.Now().Unix()))

	return m, nil
}

// RecordHTTPRequest records metrics for an HTTP request
func (m *Metrics) RecordHTTPRequest(ctx context.Context, method, endpoint string, statusCode int, duration time.Duration) {
	labels := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
		attribute.String("status_code", strconv.Itoa(statusCode)),
	)

	m.HTTPRequestsTotal.Add(ctx, 1, labels)
	m.HTTPRequestDuration.Record(ctx, duration.Seconds(), labels)
}

// IncHTTPRequestsInFlight increments the in-flight HTTP requests counter
func (m *Metrics) IncHTTPRequestsInFlight(ctx context.Context, method, endpoint string) {
	labels := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
	)
	m.HTTPRequestsInFlight.Add(ctx, 1, labels)
}

// DecHTTPRequestsInFlight decrements the in-flight HTTP requests counter
func (m *Metrics) DecHTTPRequestsInFlight(ctx context.Context, method, endpoint string) {
	labels := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
	)
	m.HTTPRequestsInFlight.Add(ctx, -1, labels)
}

// RecordAPICall records metrics for an external API call
func (m *Metrics) RecordAPICall(ctx context.Context, api, operation string, success bool, duration time.Duration) {
	labels := metric.WithAttributes(
		attribute.String("api", api),
		attribute.String("operation", operation),
		attribute.Bool("success", success),
	)

	m.APICallsTotal.Add(ctx, 1, labels)
	m.APICallDuration.Record(ctx, duration.Seconds(), labels)

	if !success {
		failureLabels := metric.WithAttributes(
			attribute.String("api", api),
			attribute.String("operation", operation),
		)
		m.APICallsFailures.Add(ctx, 1, failureLabels)
	}
}

// RecordCalendarGenerated records when a calendar is generated
func (m *Metrics) RecordCalendarGenerated(ctx context.Context, billID string, numOutages int) {
	labels := metric.WithAttributes(
		attribute.String("bill_id_prefix", billIDPrefix(billID)),
		attribute.Int("num_outages", numOutages),
	)
	m.CalendarGenerated.Add(ctx, 1, labels)
}

// RecordOutagesFetched records when outages are fetched
func (m *Metrics) RecordOutagesFetched(ctx context.Context, billID string, count int) {
	labels := metric.WithAttributes(
		attribute.String("bill_id_prefix", billIDPrefix(billID)),
		attribute.Int("count", count),
	)
	m.OutagesFetched.Add(ctx, 1, labels)
}

// billIDPrefix returns the first 3 characters of bill ID for anonymization
func billIDPrefix(billID string) string {
	if len(billID) >= 3 {
		return billID[:3] + "***"
	}
	return "***"
}
