package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/MoKhajavi75/barghvim/internal/calendar"
	"github.com/MoKhajavi75/barghvim/internal/outages"
	"github.com/MoKhajavi75/barghvim/pkg/logging"
	"github.com/MoKhajavi75/barghvim/pkg/metrics"
	"github.com/MoKhajavi75/barghvim/pkg/telemetry"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
)

func New(tel *telemetry.Telemetry) *gin.Engine {
	r := gin.New()

	// Set up metrics
	appMetrics, err := metrics.New(tel.Meter)
	if err != nil {
		logging.Error(context.Background(), "Failed to initialize metrics", err)
		return nil
	}

	// Add OpenTelemetry middleware
	r.Use(otelgin.Middleware("barghvim"))

	// Add custom middleware for structured logging and metrics
	r.Use(loggingMiddleware())
	r.Use(metricsMiddleware(appMetrics))
	r.Use(correlationIDMiddleware())

	// Add Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/:bill/cal.ics", calendarHandler(tel, appMetrics))
	}

	return r
}

func calendarHandler(tel *telemetry.Telemetry, appMetrics *metrics.Metrics) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Start tracing span
		spanCtx, span := tel.Tracer.Start(ctx.Request.Context(), "calendar_handler")
		defer span.End()

		// Update context with span
		ctx.Request = ctx.Request.WithContext(spanCtx)

		bill := ctx.Param("bill")
		token := ctx.Query("token")

		// Add attributes to span
		span.SetAttributes(
			attribute.String("bill.id", bill),
			attribute.Bool("token.provided", token != ""),
		)

		if bill == "" {
			span.RecordError(errors.New("missing bill parameter"))
			logging.Warn(spanCtx, "Missing bill parameter in request")
			ctx.String(http.StatusBadRequest, "missing bill")
			return
		}

		if token == "" {
			span.RecordError(errors.New("missing token parameter"))
			logging.Warn(spanCtx, "Missing token parameter in request")
			ctx.String(http.StatusBadRequest, "missing token")
			return
		}

		logging.Infof(spanCtx, "Processing calendar request for bill %s", bill)

		// Fetch outages with tracing
		outagesList, err := outages.Fetch(spanCtx, token, bill)
		if err != nil {
			span.RecordError(err)
			logging.Error(spanCtx, "Failed to fetch outages", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to fetch outages",
				"details": err.Error(),
			})
			return
		}

		// Record metrics for outages fetched
		appMetrics.RecordOutagesFetched(spanCtx, bill, len(outagesList))
		logging.Infof(spanCtx, "Fetched %d outages for bill %s", len(outagesList), bill)

		// Build calendar with tracing
		icsBytes, err := calendar.BuildICSWithContext(spanCtx, bill, outagesList)
		if err != nil {
			span.RecordError(err)
			logging.Error(spanCtx, "Failed to build calendar", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to build calendar",
				"details": err.Error(),
			})
			return
		}

		// Record metrics for calendar generated
		appMetrics.RecordCalendarGenerated(spanCtx, bill, len(outagesList))

		// Add span attributes for response
		span.SetAttributes(
			attribute.Int("outages.count", len(outagesList)),
			attribute.Int("calendar.size_bytes", len(icsBytes)),
		)

		ctx.Header("Content-Type", "text/calendar; charset=utf-8")
		ctx.Header("Cache-Control", "no-store")
		ctx.Header("Pragma", "no-cache")
		ctx.Header("Expires", time.Unix(0, 0).UTC().Format(http.TimeFormat))

		logging.Infof(spanCtx, "Successfully generated calendar for bill %s with %d outages", bill, len(outagesList))
		ctx.Data(http.StatusOK, "text/calendar; charset=utf-8", icsBytes)
	}
}

// correlationIDMiddleware adds a correlation ID to each request
func correlationIDMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		correlationID := telemetry.CorrelationID()

		// Add to context
		newCtx := context.WithValue(ctx.Request.Context(), logging.CorrelationIDKey, correlationID)
		ctx.Request = ctx.Request.WithContext(newCtx)

		// Add to response headers
		ctx.Header("X-Correlation-ID", correlationID)

		ctx.Next()
	}
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		// Process request
		ctx.Next()

		// Log request
		duration := time.Since(start).Seconds()
		logging.LogHTTPRequest(
			ctx.Request.Context(),
			ctx.Request.Method,
			ctx.Request.URL.Path,
			ctx.Writer.Status(),
			duration,
		)
	}
}

// metricsMiddleware records metrics for all HTTP requests
func metricsMiddleware(appMetrics *metrics.Metrics) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		// Increment in-flight requests
		appMetrics.IncHTTPRequestsInFlight(ctx.Request.Context(), ctx.Request.Method, ctx.Request.URL.Path)

		// Process request
		ctx.Next()

		// Decrement in-flight requests and record metrics
		duration := time.Since(start)
		appMetrics.DecHTTPRequestsInFlight(ctx.Request.Context(), ctx.Request.Method, ctx.Request.URL.Path)
		appMetrics.RecordHTTPRequest(ctx.Request.Context(), ctx.Request.Method, ctx.Request.URL.Path, ctx.Writer.Status(), duration)
	}
}
