package calendar

import (
	"context"
	"time"

	"github.com/MoKhajavi75/barghvim/internal/outages"
	"github.com/MoKhajavi75/barghvim/pkg/logging"
	"github.com/MoKhajavi75/barghvim/pkg/uid"
	ics "github.com/arran4/golang-ical"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const tzid = "Asia/Tehran"

func BuildICS(bill string, items []outages.Outage) ([]byte, error) {
	return BuildICSWithContext(context.Background(), bill, items)
}

func BuildICSWithContext(ctx context.Context, bill string, items []outages.Outage) ([]byte, error) {
	// Start tracing span
	tracer := otel.Tracer("calendar")
	spanCtx, span := tracer.Start(ctx, "build_ics")
	defer span.End()

	// Add span attributes
	span.SetAttributes(
		attribute.String("bill.id", bill),
		attribute.Int("outages.count", len(items)),
	)

	logging.Debugf(spanCtx, "Building ICS calendar for bill %s with %d outages", bill, len(items))

	loc, err := time.LoadLocation(tzid)
	if err != nil {
		span.RecordError(err)
		logging.Error(spanCtx, "Failed to load timezone", err)
		return nil, err
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//MoKhajavi75//barghvim 1.0.0//EN")
	cal.SetUrl("https://github.com/mokhajavi75/barghvim")
	cal.SetName("Power Outages – " + bill)
	cal.SetCalscale("GREGORIAN")

	eventsAdded := 0
	for _, o := range items {
		start := o.Start.In(loc)
		end := o.End.In(loc)

		ev := cal.AddEvent(uid.EventUID(bill, start, end))
		ev.SetSummary("Planned Power Outage")
		ev.SetTimeTransparency(ics.TransparencyTransparent)
		ev.SetStartAt(start)
		ev.SetEndAt(end)
		eventsAdded++
	}

	span.SetAttributes(
		attribute.Int("events.added", eventsAdded),
	)

	calendarBytes := []byte(cal.Serialize())
	span.SetAttributes(
		attribute.Int("calendar.size_bytes", len(calendarBytes)),
	)

	logging.Infof(spanCtx, "Successfully built ICS calendar for bill %s: %d events, %d bytes", bill, eventsAdded, len(calendarBytes))
	return calendarBytes, nil
}
