package outages

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	httpc "github.com/MoKhajavi75/barghvim/pkg/http"
)

func TestFetch(t *testing.T) {
	// Ensure gock is properly configured to intercept our HTTP client
	defer gock.Off()
	
	// Configure gock to intercept the default HTTP client used by our code
	gock.InterceptClient(httpc.Default)

	t.Run("Successful fetch", func(t *testing.T) {
		defer gock.Clean()

		// Mock successful API response
		mockResponse := respBody{
			Status:  200,
			Message: "Success",
			Data: []respItem{
				{
					OutageDate: "1403/03/25",
					StartTime:  "09:00",
					StopTime:   "12:00",
				},
				{
					OutageDate: "1403/03/26",
					StartTime:  "14:00",
					StopTime:   "16:30",
				},
			},
		}

		gock.New("https://uiapi2.saapa.ir").
			Post("/api/ebills/PlannedBlackoutsReport").
			Reply(200).
			JSON(mockResponse)

		ctx := context.Background()
		token := "test-token"
		bill := "12345678"

		outages, err := Fetch(ctx, token, bill)

		assert.NoError(t, err, "Should fetch outages successfully")
		assert.Len(t, outages, 2, "Should return correct number of outages")

		// Verify first outage
		assert.False(t, outages[0].Start.IsZero(), "Start time should be set")
		assert.False(t, outages[0].End.IsZero(), "End time should be set")
		assert.True(t, outages[0].End.After(outages[0].Start), "End should be after start")

		// Verify timezone
		loc, _ := time.LoadLocation("Asia/Tehran")
		assert.Equal(t, loc, outages[0].Start.Location(), "Should be in Tehran timezone")
		assert.Equal(t, loc, outages[0].End.Location(), "Should be in Tehran timezone")

		// Verify gock expectations
		assert.True(t, gock.IsDone(), "All HTTP expectations should be met")
	})

	t.Run("API returns error status", func(t *testing.T) {
		defer gock.Clean()

		mockResponse := respBody{
			Status:  400,
			Message: "Invalid bill ID",
			Data:    []respItem{},
		}

		gock.New("https://uiapi2.saapa.ir").
			Post("/api/ebills/PlannedBlackoutsReport").
			Reply(200).
			JSON(mockResponse)

		ctx := context.Background()
		token := "test-token"
		bill := "invalid"

		outages, err := Fetch(ctx, token, bill)

		assert.Error(t, err, "Should return error for API error status")
		assert.Nil(t, outages, "Should not return outages on error")
		assert.Contains(t, err.Error(), "upstream status 400", "Error should contain status code")
		assert.Contains(t, err.Error(), "Invalid bill ID", "Error should contain API message")
	})

	t.Run("HTTP error response", func(t *testing.T) {
		defer gock.Clean()

		gock.New("https://uiapi2.saapa.ir").
			Post("/api/ebills/PlannedBlackoutsReport").
			Reply(500).
			BodyString("Internal Server Error")

		ctx := context.Background()
		token := "test-token"
		bill := "12345678"

		outages, err := Fetch(ctx, token, bill)

		assert.Error(t, err, "Should return error for HTTP error")
		assert.Nil(t, outages, "Should not return outages on error")
		assert.Contains(t, err.Error(), "upstream http 500", "Error should contain HTTP status")
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		defer gock.Clean()

		gock.New("https://uiapi2.saapa.ir").
			Post("/api/ebills/PlannedBlackoutsReport").
			Reply(200).
			BodyString("invalid json {")

		ctx := context.Background()
		token := "test-token"
		bill := "12345678"

		outages, err := Fetch(ctx, token, bill)

		assert.Error(t, err, "Should return error for invalid JSON")
		assert.Nil(t, outages, "Should not return outages on error")
	})

	t.Run("Invalid date format in response", func(t *testing.T) {
		defer gock.Clean()

		mockResponse := respBody{
			Status:  200,
			Message: "Success",
			Data: []respItem{
				{
					OutageDate: "invalid-date",
					StartTime:  "09:00",
					StopTime:   "12:00",
				},
			},
		}

		gock.New("https://uiapi2.saapa.ir").
			Post("/api/ebills/PlannedBlackoutsReport").
			Reply(200).
			JSON(mockResponse)

		ctx := context.Background()
		token := "test-token"
		bill := "12345678"

		outages, err := Fetch(ctx, token, bill)

		assert.Error(t, err, "Should return error for invalid date format")
		assert.Nil(t, outages, "Should not return outages on error")
		assert.Contains(t, err.Error(), "bad start time", "Error should mention date parsing issue")
	})

	t.Run("Invalid time format in response", func(t *testing.T) {
		defer gock.Clean()

		mockResponse := respBody{
			Status:  200,
			Message: "Success",
			Data: []respItem{
				{
					OutageDate: "1403/03/25",
					StartTime:  "invalid-time",
					StopTime:   "12:00",
				},
			},
		}

		gock.New("https://uiapi2.saapa.ir").
			Post("/api/ebills/PlannedBlackoutsReport").
			Reply(200).
			JSON(mockResponse)

		ctx := context.Background()
		token := "test-token"
		bill := "12345678"

		outages, err := Fetch(ctx, token, bill)

		assert.Error(t, err, "Should return error for invalid time format")
		assert.Nil(t, outages, "Should not return outages on error")
		assert.Contains(t, err.Error(), "bad start time", "Error should mention time parsing issue")
	})

	t.Run("End time before start time", func(t *testing.T) {
		defer gock.Clean()

		mockResponse := respBody{
			Status:  200,
			Message: "Success",
			Data: []respItem{
				{
					OutageDate: "1403/03/25",
					StartTime:  "12:00",
					StopTime:   "09:00", // Before start time
				},
			},
		}

		gock.New("https://uiapi2.saapa.ir").
			Post("/api/ebills/PlannedBlackoutsReport").
			Reply(200).
			JSON(mockResponse)

		ctx := context.Background()
		token := "test-token"
		bill := "12345678"

		outages, err := Fetch(ctx, token, bill)

		assert.Error(t, err, "Should return error when end is before start")
		assert.Nil(t, outages, "Should not return outages on error")
		assert.Contains(t, err.Error(), "stop before start", "Error should mention time order issue")
	})

	t.Run("Context cancellation", func(t *testing.T) {
		defer gock.Clean()

		// Don't set up any mock - we'll cancel before the request
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		token := "test-token"
		bill := "12345678"

		outages, err := Fetch(ctx, token, bill)

		assert.Error(t, err, "Should return error for cancelled context")
		assert.Nil(t, outages, "Should not return outages on error")
	})

	t.Run("Request headers and body", func(t *testing.T) {
		defer gock.Clean()

		mockResponse := respBody{
			Status:  200,
			Message: "Success",
			Data:    []respItem{},
		}

		gock.New("https://uiapi2.saapa.ir").
			Post("/api/ebills/PlannedBlackoutsReport").
			MatchHeader("Content-Type", "application/json; charset=utf-8").
			MatchHeader("Authorization", "Bearer test-token").
			Reply(200).
			JSON(mockResponse)

		ctx := context.Background()
		token := "test-token"
		bill := "12345678"

		_, err := Fetch(ctx, token, bill)

		assert.NoError(t, err, "Should complete request successfully")
		assert.True(t, gock.IsDone(), "Should match all request expectations")
	})
}

func TestFetch_RequestBodyValidation(t *testing.T) {
	defer gock.Off()
	defer gock.Clean()

	// Capture the request body
	var requestBody map[string]interface{}
	
	gock.New("https://uiapi2.saapa.ir").
		Post("/api/ebills/PlannedBlackoutsReport").
		SetMatcher(gock.NewMatcher()).
		AddMatcher(func(req *http.Request, _ *gock.Request) (bool, error) {
			// Capture and validate request body
			decoder := json.NewDecoder(req.Body)
			err := decoder.Decode(&requestBody)
			return err == nil, err
		}).
		Reply(200).
		JSON(respBody{Status: 200, Message: "Success", Data: []respItem{}})

	ctx := context.Background()
	token := "test-token"
	bill := "87654321"

	_, err := Fetch(ctx, token, bill)
	require.NoError(t, err)

	// Validate request body structure
	assert.Equal(t, bill, requestBody["bill_id"], "Should send correct bill ID")
	assert.Contains(t, requestBody, "from_date", "Should include from_date")
	assert.Contains(t, requestBody, "to_date", "Should include to_date")

	// Validate date format (should be Jalali yyyy/mm/dd)
	fromDate, ok := requestBody["from_date"].(string)
	assert.True(t, ok, "from_date should be string")
	assert.Regexp(t, `^\d{4}/\d{2}/\d{2}$`, fromDate, "from_date should be yyyy/mm/dd format")

	toDate, ok := requestBody["to_date"].(string)
	assert.True(t, ok, "to_date should be string")
	assert.Regexp(t, `^\d{4}/\d{2}/\d{2}$`, toDate, "to_date should be yyyy/mm/dd format")
}

func TestFetch_DateRangeLogic(t *testing.T) {
	defer gock.Off()
	defer gock.Clean()

	// Test that we request the right date range (7 days from now)
	gock.New("https://uiapi2.saapa.ir").
		Post("/api/ebills/PlannedBlackoutsReport").
		Reply(200).
		JSON(respBody{Status: 200, Message: "Success", Data: []respItem{}})

	ctx := context.Background()
	_, err := Fetch(ctx, "test-token", "12345678")
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), "Should match all request expectations")
}

func TestFetch_EmptyResponse(t *testing.T) {
	defer gock.Off()
	defer gock.Clean()

	mockResponse := respBody{
		Status:  200,
		Message: "No outages found",
		Data:    []respItem{}, // Empty array
	}

	gock.New("https://uiapi2.saapa.ir").
		Post("/api/ebills/PlannedBlackoutsReport").
		Reply(200).
		JSON(mockResponse)

	ctx := context.Background()
	outages, err := Fetch(ctx, "test-token", "12345678")

	assert.NoError(t, err, "Should handle empty response successfully")
	assert.NotNil(t, outages, "Should return non-nil slice")
	assert.Len(t, outages, 0, "Should return empty slice")
}

func TestFetch_NetworkError(t *testing.T) {
	defer gock.Off()
	defer gock.Clean()

	gock.New("https://uiapi2.saapa.ir").
		Post("/api/ebills/PlannedBlackoutsReport").
		ReplyError(assert.AnError)

	ctx := context.Background()
	outages, err := Fetch(ctx, "test-token", "12345678")

	assert.Error(t, err, "Should return error for network issues")
	assert.Nil(t, outages, "Should not return outages on network error")
}

// Benchmark tests
func BenchmarkFetch_MockedResponse(b *testing.B) {
	defer gock.Off()
	
	// Configure gock to intercept the default HTTP client
	gock.InterceptClient(httpc.Default)
	
	mockResponse := respBody{
		Status:  200,
		Message: "Success",
		Data: []respItem{
			{OutageDate: "1403/03/25", StartTime: "09:00", StopTime: "12:00"},
			{OutageDate: "1403/03/26", StartTime: "14:00", StopTime: "16:30"},
		},
	}

	// Set up persistent mock for benchmark
	gock.New("https://uiapi2.saapa.ir").
		Post("/api/ebills/PlannedBlackoutsReport").
		Persist().
		Reply(200).
		JSON(mockResponse)

	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Fetch(ctx, "test-token", "12345678")
	}
	
	gock.Clean()
}
