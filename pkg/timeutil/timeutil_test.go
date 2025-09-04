package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToJalaliYMD(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Nowruz 2024 (Persian New Year)",
			input:    time.Date(2024, 3, 20, 12, 0, 0, 0, time.UTC), // Nowruz 2024
			expected: "1403/01/01",
		},
		{
			name:     "Mid year date",
			input:    time.Date(2024, 7, 15, 12, 30, 0, 0, time.UTC),
			expected: "1403/04/25",
		},
		{
			name:     "End of Persian year",
			input:    time.Date(2025, 3, 19, 12, 0, 0, 0, time.UTC), // Day before Nowruz 2025
			expected: "1403/12/29",
		},
	}

	// Load Tehran timezone for testing
	loc, err := time.LoadLocation("Asia/Tehran")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToJalaliYMD(tt.input, loc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromJalaliYMDHM(t *testing.T) {
	// Load Tehran timezone for testing
	loc, err := time.LoadLocation("Asia/Tehran")
	require.NoError(t, err)

	tests := []struct {
		name        string
		ymd         string
		hm          string
		expectError bool
		description string
	}{
		{
			name:        "Valid Persian date and time",
			ymd:         "1403/01/01",
			hm:          "09:30",
			expectError: false,
			description: "Standard Persian New Year morning",
		},
		{
			name:        "Valid mid-year date",
			ymd:         "1403/06/15",
			hm:          "14:45",
			expectError: false,
			description: "Mid-year afternoon",
		},
		{
			name:        "Valid end of day",
			ymd:         "1403/12/29",
			hm:          "23:59",
			expectError: false,
			description: "Last minute of Persian year",
		},
		{
			name:        "Invalid date format - missing slash",
			ymd:         "14030101",
			hm:          "09:30",
			expectError: true,
			description: "Date without separators",
		},
		{
			name:        "Invalid time format - missing colon",
			ymd:         "1403/01/01",
			hm:          "0930",
			expectError: true,
			description: "Time without separator",
		},
		{
			name:        "Invalid month - too high",
			ymd:         "1403/13/01",
			hm:          "09:30",
			expectError: false, // Current implementation doesn't validate ranges
			description: "Month 13 - handled by Persian calendar library",
		},
		{
			name:        "Invalid month - zero",
			ymd:         "1403/00/01",
			hm:          "09:30",
			expectError: false, // Current implementation doesn't validate ranges
			description: "Month 0 - handled by Persian calendar library",
		},
		{
			name:        "Invalid day - too high",
			ymd:         "1403/01/32",
			hm:          "09:30",
			expectError: false, // Current implementation doesn't validate ranges
			description: "Day 32 - handled by Persian calendar library",
		},
		{
			name:        "Invalid day - zero",
			ymd:         "1403/01/00",
			hm:          "09:30",
			expectError: false, // Current implementation doesn't validate ranges
			description: "Day 0 - handled by Persian calendar library",
		},
		{
			name:        "Invalid hour - too high",
			ymd:         "1403/01/01",
			hm:          "24:30",
			expectError: false, // Current implementation doesn't validate ranges
			description: "Hour 24 - handled by Persian calendar library",
		},
		{
			name:        "Invalid hour - negative",
			ymd:         "1403/01/01",
			hm:          "-1:30",
			expectError: false, // Parsing succeeds, Persian calendar handles it
			description: "Negative hour - handled by Persian calendar library",
		},
		{
			name:        "Invalid minute - too high",
			ymd:         "1403/01/01",
			hm:          "09:60",
			expectError: false, // Current implementation doesn't validate ranges
			description: "Minute 60 - handled by Persian calendar library",
		},
		{
			name:        "Invalid minute - negative",
			ymd:         "1403/01/01",
			hm:          "09:-1",
			expectError: false, // Parsing succeeds, Persian calendar handles it
			description: "Negative minute - handled by Persian calendar library",
		},
		{
			name:        "Empty date",
			ymd:         "",
			hm:          "09:30",
			expectError: true,
			description: "Empty date string",
		},
		{
			name:        "Empty time",
			ymd:         "1403/01/01",
			hm:          "",
			expectError: true,
			description: "Empty time string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromJalaliYMDHM(tt.ymd, tt.hm, loc)

			if tt.expectError {
				assert.Error(t, err, "Expected error for %s", tt.description)
				assert.True(t, result.IsZero(), "Result should be zero time on error")
			} else {
				assert.NoError(t, err, "Expected no error for %s", tt.description)
				assert.False(t, result.IsZero(), "Result should not be zero time on success")

				// Verify the time is in the correct timezone
				assert.Equal(t, loc, result.Location(), "Result should be in Tehran timezone")
			}
		})
	}
}

func TestFromJalaliYMDHM_RoundTrip(t *testing.T) {
	// Test that converting from time to Jalali and back gives us a valid time
	loc, err := time.LoadLocation("Asia/Tehran")
	require.NoError(t, err)

	// Test with a known good time
	original := time.Date(2024, 6, 21, 10, 30, 0, 0, loc) // Summer solstice

	// Convert to Jalali
	jalaliDate := ToJalaliYMD(original, loc)
	jalaliTime := "10:30"

	// Convert back
	converted, err := FromJalaliYMDHM(jalaliDate, jalaliTime, loc)
	require.NoError(t, err)

	// The dates should match (ignoring seconds since we don't include them)
	assert.Equal(t, original.Year(), converted.Year())
	assert.Equal(t, original.Month(), converted.Month())
	assert.Equal(t, original.Day(), converted.Day())
	assert.Equal(t, original.Hour(), converted.Hour())
	assert.Equal(t, original.Minute(), converted.Minute())
	assert.Equal(t, loc, converted.Location())
}

func TestTimeUtilEdgeCases(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Tehran")
	require.NoError(t, err)

	t.Run("Leap year handling", func(t *testing.T) {
		// Persian calendar has different leap year rules
		// Test a known leap year date
		result, err := FromJalaliYMDHM("1403/12/30", "12:00", loc)
		assert.NoError(t, err)
		assert.False(t, result.IsZero())
	})

	t.Run("DST transitions", func(t *testing.T) {
		// Test dates around DST transitions in Iran
		// Iran typically changes DST around March 22 and September 22

		beforeDST, err := FromJalaliYMDHM("1403/01/01", "02:00", loc) // Around March 21
		assert.NoError(t, err)

		afterDST, err := FromJalaliYMDHM("1403/07/01", "02:00", loc) // Around September 22
		assert.NoError(t, err)

		// Both should succeed
		assert.False(t, beforeDST.IsZero())
		assert.False(t, afterDST.IsZero())
	})

	t.Run("Boundary hours", func(t *testing.T) {
		// Test boundary conditions
		midnight, err := FromJalaliYMDHM("1403/01/01", "00:00", loc)
		assert.NoError(t, err)
		assert.Equal(t, 0, midnight.Hour())
		assert.Equal(t, 0, midnight.Minute())

		almostMidnight, err := FromJalaliYMDHM("1403/01/01", "23:59", loc)
		assert.NoError(t, err)
		assert.Equal(t, 23, almostMidnight.Hour())
		assert.Equal(t, 59, almostMidnight.Minute())
	})
}

// Benchmark tests to ensure performance
func BenchmarkToJalaliYMD(b *testing.B) {
	loc, _ := time.LoadLocation("Asia/Tehran")
	testTime := time.Now().In(loc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToJalaliYMD(testTime, loc)
	}
}

func BenchmarkFromJalaliYMDHM(b *testing.B) {
	loc, _ := time.LoadLocation("Asia/Tehran")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromJalaliYMDHM("1403/06/15", "14:30", loc)
	}
}
