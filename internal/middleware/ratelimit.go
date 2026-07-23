// Package middleware contains middlewares for api
package middleware

import (
	"net/http"
	"sync"
	"time"
)

type Visitor struct {
	tokens   float64
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*Visitor
	rate     float64
	capacity float64
}

func NewRateLimiter(rate, capacity float64) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		capacity: capacity,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &Visitor{
			tokens:   rl.capacity,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}

	now := time.Now()
	elapsed := now.Sub(v.lastSeen).Seconds()

	v.tokens += elapsed * rl.rate

	if v.tokens > rl.capacity {
		v.tokens = rl.capacity
	}

	if v.tokens >= 1.0 {
		v.tokens -= 1.0
		v.lastSeen = now
		return true
	}

	return false
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		if !rl.Allow(ip) {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
