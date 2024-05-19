package rate_limiter

import (
	"net/http"
	"sync"
	"time"
)

const (
	transactionRequest = "transaction"
	messageRequest     = "message"
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
		requestType := req.Header.Get("X-Request-Type")

		if ok, code := r.allowRequest(ip, userID, requestType); !ok {
			http.Error(w, "Rate limit exceeded", code)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (r *RateLimiter) allowRequest(ip, userID, requestType string) (bool, int) {
	now := time.Now()

	ipCount, _ := r.ipRequests.LoadOrStore(ip, &RateCounter{lastTime: now})
	ipCounter := ipCount.(*RateCounter)

	if !ipCounter.allow(r.config.MaxRequestsPerMin, time.Minute) {
		return false, http.StatusTooManyRequests
	}

	switch requestType {
	case messageRequest:
		messageCount, _ := r.userMessages.LoadOrStore(userID, &RateCounter{lastTime: now})
		messageCounter := messageCount.(*RateCounter)

		if !messageCounter.allow(r.config.MaxMessagesPerSec, time.Second) {
			return false, http.StatusTooManyRequests
		}
	case transactionRequest:
		userFailureCount, _ := r.userFailures.LoadOrStore(userID, 0)
		r.userFailures.Store(userID, userFailureCount.(int)+1)

		if userFailureCount.(int) >= r.config.MaxFailedTransactionsPerDay {
			return false, http.StatusTooManyRequests
		}

		lastFailure, _ := r.userLastFailure.LoadOrStore(userID, now)

		if now.Sub(lastFailure.(time.Time)) >= 24*time.Hour {
			r.userFailures.Store(userID, 0)
		}
	}

	return true, http.StatusOK
}

func (r *RateLimiter) resetCounters() {
	r.userFailures = sync.Map{}
	r.ipRequests = sync.Map{}
	r.userMessages = sync.Map{}
	r.userLastFailure = sync.Map{}
}
