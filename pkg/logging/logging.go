package logging

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// Logger is the structured logger instance
var Logger *logrus.Logger

// ContextKey represents the type of keys used in context
type ContextKey string

const (
	// CorrelationIDKey is the key used to store correlation ID in context
	CorrelationIDKey ContextKey = "correlation_id"
	// RequestIDKey is the key used to store request ID in context
	RequestIDKey ContextKey = "request_id"
)

// init initializes the global logger
func init() {
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(logrus.InfoLevel)
}

// SetLevel sets the logging level
func SetLevel(level logrus.Level) {
	Logger.SetLevel(level)
}

// WithContext creates a logger entry with context information
func WithContext(ctx context.Context) *logrus.Entry {
	entry := Logger.WithContext(ctx)

	// Add correlation ID if present
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		entry = entry.WithField("correlation_id", correlationID)
	}

	// Add request ID if present
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		entry = entry.WithField("request_id", requestID)
	}

	// Add trace information if available
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		entry = entry.WithFields(logrus.Fields{
			"trace_id": span.SpanContext().TraceID().String(),
			"span_id":  span.SpanContext().SpanID().String(),
		})
	}

	return entry
}

// WithFields creates a logger entry with additional fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

// WithError creates a logger entry with an error
func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}

// Info logs an info message
func Info(ctx context.Context, msg string) {
	WithContext(ctx).Info(msg)
}

// Infof logs an info message with formatting
func Infof(ctx context.Context, format string, args ...interface{}) {
	WithContext(ctx).Infof(format, args...)
}

// Error logs an error message
func Error(ctx context.Context, msg string, err error) {
	WithContext(ctx).WithError(err).Error(msg)
}

// Errorf logs an error message with formatting
func Errorf(ctx context.Context, err error, format string, args ...interface{}) {
	WithContext(ctx).WithError(err).Errorf(format, args...)
}

// Warn logs a warning message
func Warn(ctx context.Context, msg string) {
	WithContext(ctx).Warn(msg)
}

// Warnf logs a warning message with formatting
func Warnf(ctx context.Context, format string, args ...interface{}) {
	WithContext(ctx).Warnf(format, args...)
}

// Debug logs a debug message
func Debug(ctx context.Context, msg string) {
	WithContext(ctx).Debug(msg)
}

// Debugf logs a debug message with formatting
func Debugf(ctx context.Context, format string, args ...interface{}) {
	WithContext(ctx).Debugf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(ctx context.Context, msg string, err error) {
	WithContext(ctx).WithError(err).Fatal(msg)
}

// Fatalf logs a fatal message with formatting and exits
func Fatalf(ctx context.Context, err error, format string, args ...interface{}) {
	WithContext(ctx).WithError(err).Fatalf(format, args...)
}

// LogHTTPRequest logs HTTP request information
func LogHTTPRequest(ctx context.Context, method, url string, statusCode int, duration float64) {
	WithContext(ctx).WithFields(logrus.Fields{
		"http_method":      method,
		"http_url":         url,
		"http_status_code": statusCode,
		"duration_seconds": duration,
		"event_type":       "http_request",
	}).Info("HTTP request completed")
}

// LogAPICall logs external API call information
func LogAPICall(ctx context.Context, api, operation string, success bool, duration float64, err error) {
	fields := logrus.Fields{
		"api_name":         api,
		"operation":        operation,
		"success":          success,
		"duration_seconds": duration,
		"event_type":       "api_call",
	}

	entry := WithContext(ctx).WithFields(fields)
	
	if err != nil {
		entry = entry.WithError(err)
		entry.Error("External API call failed")
	} else {
		entry.Info("External API call completed")
	}
}

// LogServiceStart logs service startup information
func LogServiceStart(ctx context.Context, serviceName, version string, port int) {
	WithContext(ctx).WithFields(logrus.Fields{
		"service_name":    serviceName,
		"service_version": version,
		"port":            port,
		"event_type":      "service_start",
	}).Info("Service started successfully")
}

// LogServiceStop logs service shutdown information
func LogServiceStop(ctx context.Context, serviceName string) {
	WithContext(ctx).WithFields(logrus.Fields{
		"service_name": serviceName,
		"event_type":   "service_stop",
	}).Info("Service stopped")
}
