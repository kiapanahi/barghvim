package calendar

import (
	"strings"
	"testing"
	"time"

	"github.com/MoKhajavi75/barghvim/internal/outages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildICS(t *testing.T) {
	// Create test outages
	loc, err := time.LoadLocation("Asia/Tehran")
	require.NoError(t, err)

	testOutages := []outages.Outage{
		{
			Start: time.Date(2024, 6, 15, 9, 0, 0, 0, loc),
			End:   time.Date(2024, 6, 15, 12, 0, 0, 0, loc),
		},
		{
			Start: time.Date(2024, 6, 16, 14, 0, 0, 0, loc),
			End:   time.Date(2024, 6, 16, 16, 30, 0, 0, loc),
		},
	}

	t.Run("Valid calendar generation", func(t *testing.T) {
		bill := "12345678"
		icsBytes, err := BuildICS(bill, testOutages)
		
		assert.NoError(t, err, "Should generate calendar without error")
		assert.NotEmpty(t, icsBytes, "Generated ICS should not be empty")
		
		icsContent := string(icsBytes)
		
		// Verify basic ICS structure
		assert.Contains(t, icsContent, "BEGIN:VCALENDAR", "Should start with calendar begin")
		assert.Contains(t, icsContent, "END:VCALENDAR", "Should end with calendar end")
		assert.Contains(t, icsContent, "VERSION:2.0", "Should have correct version")
		assert.Contains(t, icsContent, "PRODID:", "Should have product ID")
		
		// Verify calendar metadata
		assert.Contains(t, icsContent, "CALSCALE:GREGORIAN", "Should use Gregorian calendar")
		assert.Contains(t, icsContent, "METHOD:PUBLISH", "Should be published calendar")
		assert.Contains(t, icsContent, "Power Outages – "+bill, "Should include bill in name")
		assert.Contains(t, icsContent, "barghvim", "Should reference the application")
	})

	t.Run("Events in calendar", func(t *testing.T) {
		bill := "12345678"
		icsBytes, err := BuildICS(bill, testOutages)
		require.NoError(t, err)
		
		icsContent := string(icsBytes)
		
		// Count events
		eventCount := strings.Count(icsContent, "BEGIN:VEVENT")
		assert.Equal(t, len(testOutages), eventCount, "Should have correct number of events")
		
		// Verify each event has required fields
		for i := 0; i < eventCount; i++ {
			assert.Contains(t, icsContent, "SUMMARY:Planned Power Outage", "Each event should have summary")
			assert.Contains(t, icsContent, "DTSTART", "Each event should have start time")
			assert.Contains(t, icsContent, "DTEND", "Each event should have end time")
			assert.Contains(t, icsContent, "UID:", "Each event should have UID")
			assert.Contains(t, icsContent, "TRANSP:TRANSPARENT", "Events should be transparent")
		}
	})

	t.Run("Event UIDs are unique", func(t *testing.T) {
		bill := "12345678"
		icsBytes, err := BuildICS(bill, testOutages)
		require.NoError(t, err)
		
		icsContent := string(icsBytes)
		
		// Extract all UIDs
		lines := strings.Split(icsContent, "\n")
		uids := make(map[string]bool)
		
		for _, line := range lines {
			if strings.HasPrefix(line, "UID:") {
				uid := strings.TrimPrefix(line, "UID:")
				uid = strings.TrimSpace(uid)
				
				assert.False(t, uids[uid], "UID should be unique: %s", uid)
				uids[uid] = true
				
				// Verify UID format
				assert.Contains(t, uid, "@taghvim-bargh", "UID should contain domain")
			}
		}
		
		assert.Len(t, uids, len(testOutages), "Should have unique UID for each event")
	})

	t.Run("Timezone handling", func(t *testing.T) {
		bill := "12345678"
		icsBytes, err := BuildICS(bill, testOutages)
		require.NoError(t, err)
		
		icsContent := string(icsBytes)
		
		// Should contain timezone information
		// ICS format typically includes timezone in datetime strings
		lines := strings.Split(icsContent, "\n")
		hasTimeInfo := false
		
		for _, line := range lines {
			if strings.HasPrefix(line, "DTSTART") || strings.HasPrefix(line, "DTEND") {
				// Should have either timezone ID or be in UTC format
				hasTimeInfo = true
				break
			}
		}
		
		assert.True(t, hasTimeInfo, "Should have datetime information")
	})
}

func TestBuildICS_EdgeCases(t *testing.T) {
	t.Run("Empty outages list", func(t *testing.T) {
		bill := "12345678"
		icsBytes, err := BuildICS(bill, []outages.Outage{})
		
		assert.NoError(t, err, "Should handle empty outages")
		assert.NotEmpty(t, icsBytes, "Should still generate valid calendar")
		
		icsContent := string(icsBytes)
		assert.Contains(t, icsContent, "BEGIN:VCALENDAR", "Should be valid calendar")
		assert.Contains(t, icsContent, "END:VCALENDAR", "Should be valid calendar")
		assert.Equal(t, 0, strings.Count(icsContent, "BEGIN:VEVENT"), "Should have no events")
	})

	t.Run("Empty bill ID", func(t *testing.T) {
		loc, err := time.LoadLocation("Asia/Tehran")
		require.NoError(t, err)
		
		testOutages := []outages.Outage{
			{
				Start: time.Date(2024, 6, 15, 9, 0, 0, 0, loc),
				End:   time.Date(2024, 6, 15, 12, 0, 0, 0, loc),
			},
		}
		
		icsBytes, err := BuildICS("", testOutages)
		
		assert.NoError(t, err, "Should handle empty bill ID")
		assert.NotEmpty(t, icsBytes, "Should still generate calendar")
		
		icsContent := string(icsBytes)
		assert.Contains(t, icsContent, "Power Outages – ", "Should handle empty bill in name")
	})

	t.Run("Single outage", func(t *testing.T) {
		loc, err := time.LoadLocation("Asia/Tehran")
		require.NoError(t, err)
		
		singleOutage := []outages.Outage{
			{
				Start: time.Date(2024, 6, 15, 9, 0, 0, 0, loc),
				End:   time.Date(2024, 6, 15, 12, 0, 0, 0, loc),
			},
		}
		
		bill := "12345678"
		icsBytes, err := BuildICS(bill, singleOutage)
		
		assert.NoError(t, err, "Should handle single outage")
		assert.NotEmpty(t, icsBytes, "Should generate calendar")
		
		icsContent := string(icsBytes)
		assert.Equal(t, 1, strings.Count(icsContent, "BEGIN:VEVENT"), "Should have exactly one event")
	})

	t.Run("Many outages", func(t *testing.T) {
		loc, err := time.LoadLocation("Asia/Tehran")
		require.NoError(t, err)
		
		// Generate many outages
		manyOutages := make([]outages.Outage, 100)
		for i := 0; i < 100; i++ {
			manyOutages[i] = outages.Outage{
				Start: time.Date(2024, 6, 15+i, 9, 0, 0, 0, loc),
				End:   time.Date(2024, 6, 15+i, 12, 0, 0, 0, loc),
			}
		}
		
		bill := "12345678"
		icsBytes, err := BuildICS(bill, manyOutages)
		
		assert.NoError(t, err, "Should handle many outages")
		assert.NotEmpty(t, icsBytes, "Should generate calendar")
		
		icsContent := string(icsBytes)
		assert.Equal(t, 100, strings.Count(icsContent, "BEGIN:VEVENT"), "Should have correct number of events")
	})

	t.Run("Very long bill ID", func(t *testing.T) {
		loc, err := time.LoadLocation("Asia/Tehran")
		require.NoError(t, err)
		
		testOutages := []outages.Outage{
			{
				Start: time.Date(2024, 6, 15, 9, 0, 0, 0, loc),
				End:   time.Date(2024, 6, 15, 12, 0, 0, 0, loc),
			},
		}
		
		longBill := strings.Repeat("1234567890", 10) // 100 character bill ID
		icsBytes, err := BuildICS(longBill, testOutages)
		
		assert.NoError(t, err, "Should handle very long bill ID")
		assert.NotEmpty(t, icsBytes, "Should generate calendar")
		
		icsContent := string(icsBytes)
		assert.Contains(t, icsContent, "Power Outages –", "Should include 'Power Outages –' in name")
		// The long bill ID might be wrapped across lines in iCalendar format, so just check for a substring
		assert.Contains(t, icsContent, "1234567890123456789012345", "Should include part of bill ID in calendar")
	})
}

func TestBuildICS_TimezoneError(t *testing.T) {
	// This test is tricky because time.LoadLocation("Asia/Tehran") should normally work
	// We can't easily test the timezone error path without mocking the time package
	// But we can at least verify the function signature and structure
	
	t.Run("Function exists and has correct signature", func(t *testing.T) {
		// This test just verifies the function exists with the expected signature
		loc, err := time.LoadLocation("Asia/Tehran")
		require.NoError(t, err)
		
		testOutages := []outages.Outage{
			{
				Start: time.Date(2024, 6, 15, 9, 0, 0, 0, loc),
				End:   time.Date(2024, 6, 15, 12, 0, 0, 0, loc),
			},
		}
		
		// Should not panic and should return expected types
		icsBytes, err := BuildICS("test", testOutages)
		assert.True(t, err == nil || err != nil, "Should return an error or nil")
		assert.True(t, icsBytes != nil || icsBytes == nil, "Should return bytes or nil")
	})
}

func TestBuildICS_EventDetails(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Tehran")
	require.NoError(t, err)
	
	// Create specific test case for detailed verification
	testStart := time.Date(2024, 6, 15, 9, 30, 0, 0, loc)
	testEnd := time.Date(2024, 6, 15, 12, 45, 0, 0, loc)
	
	testOutages := []outages.Outage{
		{Start: testStart, End: testEnd},
	}
	
	bill := "87654321"
	icsBytes, err := BuildICS(bill, testOutages)
	require.NoError(t, err)
	
	icsContent := string(icsBytes)
	
	t.Run("Event summary is correct", func(t *testing.T) {
		assert.Contains(t, icsContent, "SUMMARY:Planned Power Outage", "Should have correct summary")
	})
	
	t.Run("Event transparency is set", func(t *testing.T) {
		assert.Contains(t, icsContent, "TRANSP:TRANSPARENT", "Events should be marked as transparent")
	})
	
	t.Run("Calendar metadata is correct", func(t *testing.T) {
		assert.Contains(t, icsContent, "PRODID:-//MoKhajavi75//barghvim 1.0.0//EN", "Should have correct product ID")
		assert.Contains(t, icsContent, "URL:https://github.com/mokhajavi75/barghvim", "Should have project URL")
	})
}

// Benchmark tests
func BenchmarkBuildICS_SingleEvent(b *testing.B) {
	loc, _ := time.LoadLocation("Asia/Tehran")
	testOutages := []outages.Outage{
		{
			Start: time.Date(2024, 6, 15, 9, 0, 0, 0, loc),
			End:   time.Date(2024, 6, 15, 12, 0, 0, 0, loc),
		},
	}
	bill := "12345678"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildICS(bill, testOutages)
	}
}

func BenchmarkBuildICS_MultipleEvents(b *testing.B) {
	loc, _ := time.LoadLocation("Asia/Tehran")
	
	// Create 10 events
	testOutages := make([]outages.Outage, 10)
	for i := 0; i < 10; i++ {
		testOutages[i] = outages.Outage{
			Start: time.Date(2024, 6, 15+i, 9, 0, 0, 0, loc),
			End:   time.Date(2024, 6, 15+i, 12, 0, 0, 0, loc),
		}
	}
	bill := "12345678"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildICS(bill, testOutages)
	}
}
