package uid

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventUID(t *testing.T) {
	tests := []struct {
		name        string
		bill        string
		start       time.Time
		end         time.Time
		description string
	}{
		{
			name:        "Standard outage event",
			bill:        "12345678",
			start:       time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC),
			end:         time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC),
			description: "3-hour morning outage",
		},
		{
			name:        "Short outage event",
			bill:        "87654321",
			start:       time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			end:         time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
			description: "30-minute afternoon outage",
		},
		{
			name:        "Long bill ID",
			bill:        "123456789012",
			start:       time.Date(2024, 6, 15, 8, 0, 0, 0, time.UTC),
			end:         time.Date(2024, 6, 15, 17, 0, 0, 0, time.UTC),
			description: "Long bill ID with 9-hour outage",
		},
		{
			name:        "Minimum bill ID",
			bill:        "12345678",
			start:       time.Date(2024, 12, 31, 23, 0, 0, 0, time.UTC),
			end:         time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC),
			description: "Year-crossing outage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid := EventUID(tt.bill, tt.start, tt.end)

			// Verify UID format
			assert.Contains(t, uid, "@taghvim-bargh", "UID should contain domain suffix")
			
			// Verify UID length (SHA1 hex + "@taghvim-bargh")
			expectedLength := 40 + len("@taghvim-bargh") // SHA1 is 40 hex chars
			assert.Equal(t, expectedLength, len(uid), "UID should have correct length")
			
			// Verify it's a valid hex string before the @
			hexPart := uid[:40]
			_, err := hex.DecodeString(hexPart)
			assert.NoError(t, err, "First part should be valid hex")
			
			// Verify it ends with the correct domain
			assert.True(t, len(uid) >= len("@taghvim-bargh"), "UID should be long enough")
			assert.Equal(t, "@taghvim-bargh", uid[len(uid)-len("@taghvim-bargh"):], "UID should end with correct domain")
		})
	}
}

func TestEventUID_Deterministic(t *testing.T) {
	// Same inputs should always produce the same UID
	bill := "12345678"
	start := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	uid1 := EventUID(bill, start, end)
	uid2 := EventUID(bill, start, end)

	assert.Equal(t, uid1, uid2, "Same inputs should produce same UID")
}

func TestEventUID_Unique(t *testing.T) {
	// Different inputs should produce different UIDs
	baseTime := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)
	
	testCases := []struct {
		name  string
		bill  string
		start time.Time
		end   time.Time
	}{
		{"base case", "12345678", baseTime, baseTime.Add(3 * time.Hour)},
		{"different bill", "87654321", baseTime, baseTime.Add(3 * time.Hour)},
		{"different start", "12345678", baseTime.Add(1 * time.Hour), baseTime.Add(4 * time.Hour)},
		{"different end", "12345678", baseTime, baseTime.Add(4 * time.Hour)},
		{"different moment same tz", "12345678", baseTime.Add(1 * time.Minute), baseTime.Add(3 * time.Hour).Add(1 * time.Minute)},
	}

	uids := make(map[string]string)
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uid := EventUID(tc.bill, tc.start, tc.end)
			
			// Check if this UID was already generated
			if existingCase, exists := uids[uid]; exists {
				t.Errorf("UID collision between '%s' and '%s': %s", tc.name, existingCase, uid)
			} else {
				uids[uid] = tc.name
			}
		})
	}
}

func TestEventUID_MatchesImplementation(t *testing.T) {
	// Test that our UID matches the expected algorithm
	bill := "12345678"
	start := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	// Calculate expected UID manually
	expected := sha1.Sum(fmt.Appendf(nil, "%s|%d|%d", bill, start.Unix(), end.Unix()))
	expectedUID := hex.EncodeToString(expected[:]) + "@taghvim-bargh"

	actualUID := EventUID(bill, start, end)

	assert.Equal(t, expectedUID, actualUID, "UID should match manual calculation")
}

func TestEventUID_EdgeCases(t *testing.T) {
	t.Run("Empty bill ID", func(t *testing.T) {
		// Even with empty bill, should still generate valid UID
		start := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)
		end := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		
		uid := EventUID("", start, end)
		assert.Contains(t, uid, "@taghvim-bargh")
		assert.Equal(t, 40+len("@taghvim-bargh"), len(uid))
	})

	t.Run("Same start and end time", func(t *testing.T) {
		// Zero-duration event should still generate valid UID
		moment := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)
		
		uid := EventUID("12345678", moment, moment)
		assert.Contains(t, uid, "@taghvim-bargh")
		assert.Equal(t, 40+len("@taghvim-bargh"), len(uid))
	})

	t.Run("Very long bill ID", func(t *testing.T) {
		// Very long bill ID should still work
		longBill := "123456789012345678901234567890"
		start := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)
		end := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		
		uid := EventUID(longBill, start, end)
		assert.Contains(t, uid, "@taghvim-bargh")
		assert.Equal(t, 40+len("@taghvim-bargh"), len(uid))
	})

	t.Run("Far future dates", func(t *testing.T) {
		// Far future dates should work
		start := time.Date(2124, 6, 15, 9, 0, 0, 0, time.UTC)
		end := time.Date(2124, 6, 15, 12, 0, 0, 0, time.UTC)
		
		uid := EventUID("12345678", start, end)
		assert.Contains(t, uid, "@taghvim-bargh")
		assert.Equal(t, 40+len("@taghvim-bargh"), len(uid))
	})

	t.Run("Unix epoch", func(t *testing.T) {
		// Unix epoch should work
		start := time.Unix(0, 0)
		end := time.Unix(3600, 0) // 1 hour later
		
		uid := EventUID("12345678", start, end)
		assert.Contains(t, uid, "@taghvim-bargh")
		assert.Equal(t, 40+len("@taghvim-bargh"), len(uid))
	})
}

func TestEventUID_TimezoneHandling(t *testing.T) {
	// UIDs should be the same regardless of timezone representation
	// if the actual moment is the same
	bill := "12345678"
	
	// Same moment in different timezones
	utcTime := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)
	tehranTime := utcTime.In(time.FixedZone("IRST", 4*3600+30*60)) // UTC+4:30
	
	utcEnd := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	tehranEnd := utcEnd.In(time.FixedZone("IRST", 4*3600+30*60))
	
	uidUTC := EventUID(bill, utcTime, utcEnd)
	uidTehran := EventUID(bill, tehranTime, tehranEnd)
	
	assert.Equal(t, uidUTC, uidTehran, "UIDs should be same for same moment in different timezones")
}

// Benchmark tests
func BenchmarkEventUID(b *testing.B) {
	bill := "12345678"
	start := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EventUID(bill, start, end)
	}
}

func BenchmarkEventUID_VaryingInputs(b *testing.B) {
	bills := []string{"12345678", "87654321", "11111111", "99999999"}
	baseTime := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bill := bills[i%len(bills)]
		start := baseTime.Add(time.Duration(i) * time.Minute)
		end := start.Add(3 * time.Hour)
		EventUID(bill, start, end)
	}
}
