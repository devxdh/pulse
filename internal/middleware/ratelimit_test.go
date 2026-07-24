package middleware

import (
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	rl := NewRateLimiter(1.0, 2.0)

	baseTime := time.Now()
	mockTime := baseTime

	rl.now = func() time.Time {
		return mockTime
	}

	testCases := []struct {
		name          string
		ip            string
		timeForward   time.Duration
		expectedAllow bool
	}{
		{"First request consumes 1 token", "192.168.1.1", 0, true},
		{"Second request consumes last token", "192.168.1.1", 0, true},
		{"Third request is blocked immediately", "192.168.1.1", 0, false},
		{"Different IP is isolated and allowed", "10.0.0.1", 0, true},
		{"Wait 1 second to refill 1 token", "192.168.1.1", time.Second, true},
		{"Immediate follow up is blocked again", "192.168.1.1", 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockTime = mockTime.Add(tc.timeForward)
			result := rl.Allow(tc.ip)

			if result != tc.expectedAllow {
				t.Errorf("Expected Allow() to be %v, got %v", tc.expectedAllow, result)
			}
		})
	}
}
