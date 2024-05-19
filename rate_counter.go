package rate_limiter

import (
	"sync"
	"time"
)

type RateCounter struct {
	mu       sync.Mutex
	count    int
	lastTime time.Time
}

func (rc *RateCounter) allow(limit int, interval time.Duration) bool {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	now := time.Now()

	if now.Sub(rc.lastTime) > interval {
		rc.count = 0
		rc.lastTime = now
	}

	if rc.count < limit {
		rc.count++
		return true
	}

	return false
}
