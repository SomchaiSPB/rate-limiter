package rate_limiter

import (
	"net/http"
	"time"
)

const (
	transactionRequest = "transaction"
	messageRequest     = "message"
)

type Storage interface {
	LoadOrStore(key, value any) (actual any, loaded bool)
	Store(key, value any)
	Reset()
}

// RateLimiter main struct for storing user rate limit data
type RateLimiter struct {
	config  Config
	storage Storage
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		config:  NewConfig(),
		storage: newStorage(),
	}
}

func NewRateLimiterWithConfig(config Config, storage Storage) *RateLimiter {
	return &RateLimiter{
		config:  config,
		storage: storage,
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

	ipCount, _ := r.storage.LoadOrStore(ip, &RateCounter{lastTime: now})
	ipCounter := ipCount.(*RateCounter)

	if !ipCounter.allow(r.config.MaxRequestsPerMin, time.Minute) {
		return false, http.StatusTooManyRequests
	}

	switch requestType {
	case messageRequest:
		key := requestType + userID
		messageCount, _ := r.storage.LoadOrStore(key, &RateCounter{lastTime: now})
		messageCounter := messageCount.(*RateCounter)

		if !messageCounter.allow(r.config.MaxMessagesPerSec, time.Second) {
			return false, http.StatusTooManyRequests
		}
	case transactionRequest:
		key := requestType + userID
		userFailureCount, _ := r.storage.LoadOrStore(key, &RateCounter{lastTime: now})
		transactionCounter := userFailureCount.(*RateCounter)
		transactionCounter.count += 1
		r.storage.Store(userID, transactionCounter)

		if now.Sub(transactionCounter.lastTime) >= 24*time.Hour {
			r.storage.Store(userID, 0)
		}

		if userFailureCount.(*RateCounter).count > r.config.MaxFailedTransactionsPerDay {
			return false, http.StatusTooManyRequests
		}
	}

	return true, http.StatusOK
}

func (r *RateLimiter) resetCounters() {
	r.storage.Reset()
}
