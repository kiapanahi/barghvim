package server

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MoKhajavi75/barghvim/internal/outages"
	httpc "github.com/MoKhajavi75/barghvim/pkg/http"
)

func TestDebugGock(t *testing.T) {
	defer gock.Off()

	// Configure gock to intercept the default HTTP client
	gock.InterceptClient(httpc.Default)

	// Mock SAAPA API response
	mockResponse := map[string]interface{}{
		"status":  200,
		"message": "Success",
		"data": []map[string]interface{}{
			{
				"outage_date":       "1403/03/25",
				"outage_start_time": "09:00",
				"outage_stop_time":  "12:00",
			},
		},
	}

	gock.New("https://uiapi2.saapa.ir").
		Post("/api/ebills/PlannedBlackoutsReport").
		Reply(200).
		JSON(mockResponse)

	// Test directly calling outages.Fetch (this should work if gock is configured correctly)
	ctx := context.Background()
	result, err := outages.Fetch(ctx, "test-token", "12345678")

	require.NoError(t, err, "Should successfully fetch outages")
	require.NotNil(t, result, "Should return outages")
	assert.Len(t, result, 1, "Should return one outage")
	assert.True(t, gock.IsDone(), "All HTTP expectations should be met")
}
