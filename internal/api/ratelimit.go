package api

import (
	"sync"
	"time"
)

// comeOverCooldown is the minimum gap between "Komme vorbei" nudges to the same
// friend, so notifications can't be spammed.
const comeOverCooldown = 5 * time.Minute

// rateLimiter is a simple in-memory per-key cooldown. It resets on restart,
// which is fine for a single-instance deployment — the goal is anti-spam, not
// durable accounting.
type rateLimiter struct {
	mu       sync.Mutex
	last     map[string]time.Time
	cooldown time.Duration
}

func newRateLimiter(cooldown time.Duration) *rateLimiter {
	return &rateLimiter{last: make(map[string]time.Time), cooldown: cooldown}
}

// allow reports whether an action for key is permitted at time now, recording
// the time when it is. Stale entries are evicted to bound memory.
func (l *rateLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if t, ok := l.last[key]; ok && now.Sub(t) < l.cooldown {
		return false
	}
	l.last[key] = now
	for k, t := range l.last {
		if now.Sub(t) >= l.cooldown {
			delete(l.last, k)
		}
	}
	return true
}
