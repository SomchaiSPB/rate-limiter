package rate_limiter

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter main struct for storing user rate limit data
type RateLimiter struct {
	config          Config
	ipRequests      sync.Map
	userMessages    sync.Map
	userFailures    sync.Map
	userLastFailure sync.Map
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		config:          NewConfig(),
		ipRequests:      sync.Map{},
		userFailures:    sync.Map{},
		userLastFailure: sync.Map{},
		userMessages:    sync.Map{},
	}
}

func NewRateLimiterWithConfig(config Config) *RateLimiter {
	return &RateLimiter{
		config:          config,
		ipRequests:      sync.Map{},
		userFailures:    sync.Map{},
		userLastFailure: sync.Map{},
		userMessages:    sync.Map{},
	}
}

func (r *RateLimiter) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ip := req.RemoteAddr
		userID := req.Header.Get("X-User-ID")

		if !r.allowRequest(ip, userID) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (r *RateLimiter) allowRequest(ip, userID string) bool {
	now := time.Now()

	ipCount, _ := r.ipRequests.LoadOrStore(ip, &RateCounter{lastTime: now})
	ipCounter := ipCount.(*RateCounter)

	if !ipCounter.allow(r.config.MaxRequestsPerMin, time.Minute) {
		return false
	}

	messageCount, _ := r.userMessages.LoadOrStore(userID, &RateCounter{lastTime: now})
	messageCounter := messageCount.(*RateCounter)

	if !messageCounter.allow(r.config.MaxMessagesPerSec, time.Second) {
		return false
	}

	userFailureCount, _ := r.userFailures.LoadOrStore(userID, 0)
	r.userFailures.Store(userID, userFailureCount.(int)+1)

	if userFailureCount.(int) > r.config.MaxFailedTransactionsPerDay {
		return false
	}

	lastFailure, _ := r.userLastFailure.LoadOrStore(userID, now)

	if now.Sub(lastFailure.(time.Time)) >= 24*time.Hour {
		r.userFailures.Store(userID, 0)
	}

	return true
}

func (r *RateLimiter) ResetCounters() {
	r.userFailures = sync.Map{}
	r.ipRequests = sync.Map{}
	r.userMessages = sync.Map{}
	r.userLastFailure = sync.Map{}
}
