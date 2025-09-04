package testutil

import (
	"time"

	"github.com/MoKhajavi75/barghvim/internal/outages"
)

// TestOutages provides common test fixtures for outages
type TestOutages struct{}

// SingleOutage returns a single test outage
func (TestOutages) SingleOutage() outages.Outage {
	loc, _ := time.LoadLocation("Asia/Tehran")
	return outages.Outage{
		Start: time.Date(2024, 6, 15, 9, 0, 0, 0, loc),
		End:   time.Date(2024, 6, 15, 12, 0, 0, 0, loc),
	}
}

// MultipleOutages returns multiple test outages
func (TestOutages) MultipleOutages() []outages.Outage {
	loc, _ := time.LoadLocation("Asia/Tehran")
	return []outages.Outage{
		{
			Start: time.Date(2024, 6, 15, 9, 0, 0, 0, loc),
			End:   time.Date(2024, 6, 15, 12, 0, 0, 0, loc),
		},
		{
			Start: time.Date(2024, 6, 16, 14, 0, 0, 0, loc),
			End:   time.Date(2024, 6, 16, 16, 30, 0, 0, loc),
		},
		{
			Start: time.Date(2024, 6, 17, 8, 0, 0, 0, loc),
			End:   time.Date(2024, 6, 17, 17, 0, 0, 0, loc),
		},
	}
}

// OverlappingOutages returns outages with overlapping times (for error testing)
func (TestOutages) OverlappingOutages() []outages.Outage {
	loc, _ := time.LoadLocation("Asia/Tehran")
	return []outages.Outage{
		{
			Start: time.Date(2024, 6, 15, 9, 0, 0, 0, loc),
			End:   time.Date(2024, 6, 15, 12, 0, 0, 0, loc),
		},
		{
			Start: time.Date(2024, 6, 15, 10, 0, 0, 0, loc), // Overlaps with first
			End:   time.Date(2024, 6, 15, 13, 0, 0, 0, loc),
		},
	}
}

// ZeroDurationOutage returns an outage with same start and end time
func (TestOutages) ZeroDurationOutage() outages.Outage {
	loc, _ := time.LoadLocation("Asia/Tehran")
	moment := time.Date(2024, 6, 15, 9, 0, 0, 0, loc)
	return outages.Outage{
		Start: moment,
		End:   moment,
	}
}

// FutureOutages returns outages far in the future
func (TestOutages) FutureOutages() []outages.Outage {
	loc, _ := time.LoadLocation("Asia/Tehran")
	return []outages.Outage{
		{
			Start: time.Date(2124, 6, 15, 9, 0, 0, 0, loc),
			End:   time.Date(2124, 6, 15, 12, 0, 0, 0, loc),
		},
	}
}

// PastOutages returns outages in the past
func (TestOutages) PastOutages() []outages.Outage {
	loc, _ := time.LoadLocation("Asia/Tehran")
	return []outages.Outage{
		{
			Start: time.Date(2020, 6, 15, 9, 0, 0, 0, loc),
			End:   time.Date(2020, 6, 15, 12, 0, 0, 0, loc),
		},
	}
}

// TestBills provides common test bill IDs
type TestBills struct{}

// Valid returns valid test bill IDs
func (TestBills) Valid() []string {
	return []string{
		"12345678",
		"87654321",
		"11111111",
		"99999999",
		"123456789012", // 12 digits
	}
}

// Invalid returns invalid test bill IDs for error testing
func (TestBills) Invalid() []string {
	return []string{
		"",              // Empty
		"1234567",       // Too short
		"1234567890123", // Too long
		"abcd1234",      // Contains letters
		"1234-5678",     // Contains special chars
	}
}

// Long returns very long bill ID for edge case testing
func (TestBills) Long() string {
	return "123456789012345678901234567890" // 30 characters
}

// TestTokens provides common test tokens
type TestTokens struct{}

// Valid returns valid test tokens
func (TestTokens) Valid() []string {
	return []string{
		"test-token-12345",
		"bearer-token-abcdef",
		"api-key-9876543210",
	}
}

// Invalid returns invalid test tokens
func (TestTokens) Invalid() []string {
	return []string{
		"",      // Empty
		"short", // Too short
		"a",     // Single character
	}
}

// Long returns very long token for edge case testing
func (TestTokens) Long() string {
	token := ""
	for i := 0; i < 100; i++ {
		token += "abcdef1234567890"
	}
	return token
}

// TestDates provides common test date and time values
type TestDates struct{}

// ValidJalaliDates returns valid Jalali date strings
func (TestDates) ValidJalaliDates() []string {
	return []string{
		"1403/01/01", // Persian New Year
		"1403/06/15", // Mid year
		"1403/12/29", // End of year
		"1400/01/01", // Different year
	}
}

// InvalidJalaliDates returns invalid Jalali date strings
func (TestDates) InvalidJalaliDates() []string {
	return []string{
		"",             // Empty
		"1403/1/1",     // Single digit month/day
		"1403/13/01",   // Invalid month
		"1403/01/32",   // Invalid day
		"1403/00/01",   // Zero month
		"1403/01/00",   // Zero day
		"invalid-date", // Non-numeric
		"14030101",     // No separators
		"1403-01-01",   // Wrong separator
	}
}

// ValidTimes returns valid time strings
func (TestDates) ValidTimes() []string {
	return []string{
		"00:00", // Midnight
		"09:00", // Morning
		"12:00", // Noon
		"18:30", // Evening
		"23:59", // End of day
	}
}

// InvalidTimes returns invalid time strings
func (TestDates) InvalidTimes() []string {
	return []string{
		"",        // Empty
		"9:00",    // Single digit hour
		"09:0",    // Single digit minute
		"24:00",   // Invalid hour
		"09:60",   // Invalid minute
		"-1:00",   // Negative hour
		"09:-1",   // Negative minute
		"invalid", // Non-numeric
		"0900",    // No separator
		"09-00",   // Wrong separator
	}
}

// TestTehranLocation returns Tehran timezone for tests
func TestTehranLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		// Fallback to fixed offset if timezone data not available
		return time.FixedZone("IRST", 3*60*60+30*60) // UTC+3:30
	}
	return loc
}

// TestUTCLocation returns UTC timezone for tests
func TestUTCLocation() *time.Location {
	return time.UTC
}

// AssertTimesEqual compares two times ignoring nanoseconds (since calendar events don't include them)
func AssertTimesEqual(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day() &&
		t1.Hour() == t2.Hour() &&
		t1.Minute() == t2.Minute() &&
		t1.Location() == t2.Location()
}

// MockSAAPA provides common SAAPA API mock responses
type MockSAAPA struct{}

// SuccessResponse returns a successful SAAPA API response
func (MockSAAPA) SuccessResponse() map[string]interface{} {
	return map[string]interface{}{
		"status":  200,
		"message": "Success",
		"data": []map[string]interface{}{
			{
				"outage_date":       "1403/03/25",
				"outage_start_time": "09:00",
				"outage_stop_time":  "12:00",
			},
			{
				"outage_date":       "1403/03/26",
				"outage_start_time": "14:00",
				"outage_stop_time":  "16:30",
			},
		},
	}
}

// EmptyResponse returns a successful but empty SAAPA API response
func (MockSAAPA) EmptyResponse() map[string]interface{} {
	return map[string]interface{}{
		"status":  200,
		"message": "No outages found",
		"data":    []interface{}{},
	}
}

// ErrorResponse returns an error SAAPA API response
func (MockSAAPA) ErrorResponse(status int, message string) map[string]interface{} {
	return map[string]interface{}{
		"status":  status,
		"message": message,
		"data":    []interface{}{},
	}
}

// InvalidDateResponse returns a response with invalid date format
func (MockSAAPA) InvalidDateResponse() map[string]interface{} {
	return map[string]interface{}{
		"status":  200,
		"message": "Success",
		"data": []map[string]interface{}{
			{
				"outage_date":       "invalid-date",
				"outage_start_time": "09:00",
				"outage_stop_time":  "12:00",
			},
		},
	}
}

// InvalidTimeResponse returns a response with invalid time format
func (MockSAAPA) InvalidTimeResponse() map[string]interface{} {
	return map[string]interface{}{
		"status":  200,
		"message": "Success",
		"data": []map[string]interface{}{
			{
				"outage_date":       "1403/03/25",
				"outage_start_time": "invalid-time",
				"outage_stop_time":  "12:00",
			},
		},
	}
}

// BadTimeOrderResponse returns a response where end time is before start time
func (MockSAAPA) BadTimeOrderResponse() map[string]interface{} {
	return map[string]interface{}{
		"status":  200,
		"message": "Success",
		"data": []map[string]interface{}{
			{
				"outage_date":       "1403/03/25",
				"outage_start_time": "12:00",
				"outage_stop_time":  "09:00", // Before start time
			},
		},
	}
}
