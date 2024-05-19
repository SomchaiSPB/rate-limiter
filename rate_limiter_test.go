package rate_limiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	conf := NewConfig().
		WithMaxMessages(1).
		WithMaxRequests(10).
		WithMaxFailedTransactions(1)

	rm := NewRateLimiterWithConfig(conf)

	type test struct {
		name         string
		ip           string
		userID       string
		simulateFn   func()
		expectedCode int
	}

	tests := []test{
		{
			name:         "Allow request within rate limit",
			ip:           "192.168.1.1",
			userID:       "user1",
			simulateFn:   func() {},
			expectedCode: http.StatusOK,
		},
		{
			name:   "Deny request exceeding IP rate limit",
			ip:     "192.168.1.2",
			userID: "user2",
			simulateFn: func() {
				now := time.Now()
				ipCount, _ := rm.ipRequests.LoadOrStore("192.168.1.2", &RateCounter{lastTime: now})
				ipCounter := ipCount.(*RateCounter)
				for i := 0; i < conf.MaxRequestsPerMin; i++ {
					ipCounter.allow(conf.MaxRequestsPerMin, time.Minute)
				}
			},
			expectedCode: http.StatusTooManyRequests,
		},
		{
			name:   "Deny request exceeding user max messages limit",
			ip:     "192.168.1.3",
			userID: "user3",
			simulateFn: func() {
				now := time.Now()
				messageCount, _ := rm.userMessages.LoadOrStore("user3", &RateCounter{lastTime: now})
				messageCounter := messageCount.(*RateCounter)
				for i := 0; i <= conf.MaxMessagesPerSec; i++ {
					messageCounter.allow(conf.MaxMessagesPerSec, time.Second)
				}
			},
			expectedCode: http.StatusTooManyRequests,
		},
		{
			name:   "Deny request exceeding user failed transactions limit",
			ip:     "192.168.1.4",
			userID: "user4",
			simulateFn: func() {
				rm.userFailures.Store("user4", conf.MaxFailedTransactionsPerDay)
			},
			expectedCode: http.StatusTooManyRequests,
		},
	}

	handler := rm.RateLimiterMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rm.ResetCounters()
			tc.simulateFn()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.ip
			req.Header.Set("X-User-ID", tc.userID)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedCode {
				t.Errorf("expected status %v, got %v", tc.expectedCode, rr.Code)
			}
		})
	}
}
