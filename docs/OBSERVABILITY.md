# OpenTelemetry Observability

This document describes the comprehensive OpenTelemetry observability implementation in the barghvim service.

## Overview

The service now includes full observability with structured logging, metrics collection, and distributed tracing using OpenTelemetry standards.

## Features

### Structured Logging
- **JSON formatted logs** with contextual information
- **Correlation IDs** for request tracing across the entire request lifecycle
- **Trace/Span IDs** automatically included when tracing is enabled
- **Structured API call logging** with timing and success/failure status
- **Configurable log levels** via environment variables

### Metrics Collection
- **HTTP request metrics**: Request count, duration, and in-flight requests
- **API call metrics**: External API call count, duration, and failure rates
- **Business metrics**: Calendar generation and outage fetching counts
- **Process metrics**: Service start time and resource usage
- **Prometheus exposition** at `/metrics` endpoint

### Distributed Tracing
- **Full request tracing** from HTTP endpoint to external API calls
- **Span creation** for all major operations (fetch outages, build calendar)
- **Error recording** with proper span error status
- **Span attributes** for request parameters and response details
- **OTLP export** support for external tracing systems

## Configuration

### Environment Variables

```bash
# Logging
LOG_LEVEL=debug|info|warn|error  # Default: info

# Observability
ENVIRONMENT=development|staging|production  # Default: development
OTLP_ENDPOINT=http://your-otlp-collector:4318  # Optional: OTLP trace export endpoint

# Example for Jaeger
OTLP_ENDPOINT=http://localhost:14268/api/traces

# Example for New Relic
OTLP_ENDPOINT=https://otlp.nr-data.net:4318
```

### Production Setup

For production deployments, consider:

1. **OTLP Endpoint**: Configure `OTLP_ENDPOINT` for trace export to your observability platform
2. **Log Level**: Set `LOG_LEVEL=info` or `LOG_LEVEL=warn` to reduce log volume
3. **Environment**: Set `ENVIRONMENT=production` for proper service identification

## Endpoints

### Health Check
```
GET /health
```
Returns service health status and timestamp.

### Metrics
```
GET /metrics
```
Prometheus-compatible metrics endpoint exposing all collected metrics.

### Main API
```
GET /v1/{bill}/cal.ics?token={token}
```
Main calendar generation endpoint, fully instrumented with logging, metrics, and tracing.

## Metrics Reference

### HTTP Metrics
- `http_requests_total`: Total HTTP requests (labels: method, endpoint, status_code)
- `http_request_duration_seconds`: HTTP request duration histogram
- `http_requests_in_flight`: Current in-flight HTTP requests

### API Metrics
- `api_calls_total`: Total external API calls (labels: api, operation, success)
- `api_call_duration_seconds`: External API call duration histogram
- `api_calls_failures_total`: Total failed API calls (labels: api, operation)

### Business Metrics
- `calendars_generated_total`: Total calendars generated (labels: bill_id_prefix, num_outages)
- `outages_fetched_total`: Total outages fetched (labels: bill_id_prefix, count)

### Process Metrics
- `process_start_time_seconds`: Service start time as Unix timestamp

## Logging Structure

All logs are structured JSON with the following fields:

```json
{
  "@timestamp": "2025-09-04T12:46:01.442Z",
  "level": "info",
  "message": "HTTP request completed",
  "correlation_id": "barghvim-1756977361442593400",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "http_method": "GET",
  "http_url": "/v1/12345678/cal.ics",
  "http_status_code": 200,
  "duration_seconds": 0.245,
  "event_type": "http_request"
}
```

### Event Types
- `http_request`: HTTP request completion
- `api_call`: External API call (SAAPA)
- `service_start`: Service startup
- `service_stop`: Service shutdown

## Tracing

Spans are created for:

1. **HTTP requests** (via otelgin middleware)
2. **Calendar handler** - Main request processing
3. **Fetch outages** - External API calls  
4. **Build ICS** - Calendar generation
5. **HTTP client calls** (via otelhttp transport)

### Span Attributes

#### Calendar Handler Span
- `bill.id`: Bill identifier
- `token.provided`: Whether token was provided
- `outages.count`: Number of outages found
- `calendar.size_bytes`: Generated calendar size

#### Outages Fetch Span
- `bill.id`: Bill identifier
- `request.from_date`: Query start date (Jalali)
- `request.to_date`: Query end date (Jalali)  
- `response.status_code`: HTTP status code
- `response.api_status`: SAAPA API status
- `response.outages_count`: Number of outages returned

#### Calendar Build Span
- `bill.id`: Bill identifier
- `outages.count`: Number of outages to process
- `events.added`: Number of calendar events created
- `calendar.size_bytes`: Final calendar size

## Integration Examples

### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'barghvim'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana Dashboard Queries

```promql
# Request rate
rate(http_requests_total[5m])

# Request duration 95th percentile
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error rate
rate(http_requests_total{status_code!~"2.."}[5m]) / rate(http_requests_total[5m])

# API call success rate
rate(api_calls_total{success="true"}[5m]) / rate(api_calls_total[5m])
```

### Jaeger Integration

Set the OTLP endpoint to your Jaeger collector:
```bash
OTLP_ENDPOINT=http://jaeger-collector:14268/api/traces
```

## Development

### Mock Telemetry

For testing, the service provides mock telemetry instances that don't emit actual metrics or traces:

```go
mockTel := telemetry.NewMockTelemetry()
router := server.New(mockTel)
```

### Testing

The observability implementation is thoroughly tested:
- All existing tests pass without modification
- Structured logs appear in test output
- Mock telemetry prevents interference with test assertions

## Performance Impact

The observability implementation is designed for minimal performance impact:

- **Lazy initialization** of expensive resources
- **Efficient JSON logging** with logrus
- **OpenTelemetry SDK optimizations** with batching and sampling
- **Minimal memory allocations** in hot paths
- **Graceful degradation** if telemetry setup fails

## Troubleshooting

### Common Issues

1. **High log volume**: Adjust `LOG_LEVEL` to `warn` or `error`
2. **Missing traces**: Verify `OTLP_ENDPOINT` configuration
3. **Metrics not appearing**: Check `/metrics` endpoint accessibility
4. **Performance issues**: Consider reducing trace sampling rate

### Debug Mode

Enable debug logging to see detailed information:
```bash
LOG_LEVEL=debug ./server
```

This will show all HTTP requests, API calls, and internal operations with full context.
