package outages

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	httpc "github.com/MoKhajavi75/barghvim/pkg/http"
	"github.com/MoKhajavi75/barghvim/pkg/logging"
	"github.com/MoKhajavi75/barghvim/pkg/timeutil"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const apiURL = "https://uiapi2.saapa.ir/api/ebills/PlannedBlackoutsReport"

type Outage struct {
	Start time.Time // Asia/Tehran
	End   time.Time // Asia/Tehran
}

type reqBody struct {
	BillID string `json:"bill_id"`
	From   string `json:"from_date"` // Jalali yyyy/mm/dd
	To     string `json:"to_date"`   // Jalali yyyy/mm/dd
}

type respBody struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    []respItem `json:"data"`
}

type respItem struct {
	OutageDate string `json:"outage_date"`       // yyyy/mm/dd (Jalali)
	StartTime  string `json:"outage_start_time"` // HH:MM
	StopTime   string `json:"outage_stop_time"`  // HH:MM
}

func Fetch(ctx context.Context, token string, bill string) ([]Outage, error) {
	// Start tracing span
	tracer := otel.Tracer("outages")
	spanCtx, span := tracer.Start(ctx, "fetch_outages")
	defer span.End()
	
	// Add span attributes
	span.SetAttributes(
		attribute.String("bill.id", bill),
		attribute.Bool("token.provided", token != ""),
	)
	
	start := time.Now()
	success := false
	defer func() {
		duration := time.Since(start).Seconds()
		logging.LogAPICall(spanCtx, "SAAPA", "PlannedBlackoutsReport", success, duration, nil)
	}()

	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		span.RecordError(err)
		logging.Error(spanCtx, "Failed to load timezone", err)
		return nil, err
	}

	from := time.Now()
	to := from.Add(7 * 24 * time.Hour)

	body := reqBody{
		BillID: bill,
		From:   timeutil.ToJalaliYMD(from.In(loc), loc),
		To:     timeutil.ToJalaliYMD(to.In(loc), loc),
	}
	
	span.SetAttributes(
		attribute.String("request.from_date", body.From),
		attribute.String("request.to_date", body.To),
	)
	
	logging.Debugf(spanCtx, "Fetching outages for bill %s from %s to %s", bill, body.From, body.To)
	
	reqBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(spanCtx, http.MethodPost, apiURL, bytes.NewReader(reqBytes))
	if err != nil {
		span.RecordError(err)
		logging.Error(spanCtx, "Failed to create HTTP request", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)

	logging.Debug(spanCtx, "Making API call to SAAPA")
	resp, err := httpc.Default.Do(req)
	if err != nil {
		span.RecordError(err)
		logging.Error(spanCtx, "HTTP request failed", err)
		return nil, err
	}

	defer resp.Body.Close()
	
	span.SetAttributes(
		attribute.Int("response.status_code", resp.StatusCode),
	)

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("upstream http %d", resp.StatusCode)
		span.RecordError(err)
		logging.Errorf(spanCtx, err, "API returned non-200 status: %d", resp.StatusCode)
		return nil, err
	}

	var rb respBody
	if err := json.NewDecoder(resp.Body).Decode(&rb); err != nil {
		span.RecordError(err)
		logging.Error(spanCtx, "Failed to decode JSON response", err)
		return nil, err
	}
	
	span.SetAttributes(
		attribute.Int("response.api_status", rb.Status),
		attribute.String("response.api_message", rb.Message),
		attribute.Int("response.outages_count", len(rb.Data)),
	)

	if rb.Status != 200 {
		err := fmt.Errorf("upstream status %d: %s", rb.Status, rb.Message)
		span.RecordError(err)
		logging.Errorf(spanCtx, err, "API returned error status: %d, message: %s", rb.Status, rb.Message)
		return nil, err
	}

	logging.Infof(spanCtx, "Received %d outages from API", len(rb.Data))

	out := make([]Outage, 0, len(rb.Data))
	for i, item := range rb.Data {
		start, err := timeutil.FromJalaliYMDHM(item.OutageDate, item.StartTime, loc)
		if err != nil {
			span.RecordError(err)
			logging.Errorf(spanCtx, err, "Bad start time in item %d (%s %s)", i, item.OutageDate, item.StartTime)
			return nil, fmt.Errorf("bad start time (%s %s): %w", item.OutageDate, item.StartTime, err)
		}

		end, err := timeutil.FromJalaliYMDHM(item.OutageDate, item.StopTime, loc)
		if err != nil {
			span.RecordError(err)
			logging.Errorf(spanCtx, err, "Bad stop time in item %d (%s %s)", i, item.OutageDate, item.StopTime)
			return nil, fmt.Errorf("bad stop time (%s %s): %w", item.OutageDate, item.StopTime, err)
		}

		if !end.After(start) {
			err := fmt.Errorf("stop before start (%s %s-%s)", item.OutageDate, item.StartTime, item.StopTime)
			span.RecordError(err)
			logging.Error(spanCtx, "Invalid time range in outage data", err)
			return nil, err
		}

		out = append(out, Outage{
			Start: start,
			End:   end,
		})
	}

	success = true
	logging.Infof(spanCtx, "Successfully parsed %d outages for bill %s", len(out), bill)
	return out, nil
}
