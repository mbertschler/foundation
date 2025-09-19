package auth

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mbertschler/foundation"
)

type RateLimit struct {
	attempts     int
	lastAttempt  time.Time
	blockedUntil time.Time
}

type RateLimiter struct {
	mu            sync.RWMutex
	limits        map[string]*RateLimit
	maxAttempts   int
	window        time.Duration
	blockDuration time.Duration
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limits:        make(map[string]*RateLimit),
		maxAttempts:   5,                // 5 attempts
		window:        time.Minute,      // per minute
		blockDuration: 15 * time.Minute, // block for 15 minutes
	}

	// Clean up old entries every 10 minutes
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, limit := range rl.limits {
			// Remove entries that are past their block time and window
			if now.After(limit.blockedUntil) && now.Sub(limit.lastAttempt) > rl.window {
				delete(rl.limits, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getClientKey(r *foundation.Request, username string) string {
	// Get IP address
	ip := r.Request.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Request.Header.Get("X-Real-IP")
	}
	if ip == "" {
		host, _, _ := net.SplitHostPort(r.Request.RemoteAddr)
		ip = host
	}

	// Combine IP and username for the key
	return fmt.Sprintf("%s:%s", ip, username)
}

func (rl *RateLimiter) IsBlocked(r *foundation.Request, username string) bool {
	key := rl.getClientKey(r, username)

	rl.mu.RLock()
	limit, exists := rl.limits[key]
	rl.mu.RUnlock()

	if !exists {
		return false
	}

	now := time.Now()
	return now.Before(limit.blockedUntil)
}

func (rl *RateLimiter) RecordAttempt(r *foundation.Request, username string, success bool) {
	key := rl.getClientKey(r, username)
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.limits[key]
	if !exists {
		limit = &RateLimit{}
		rl.limits[key] = limit
	}

	// Reset if outside the window
	if now.Sub(limit.lastAttempt) > rl.window {
		limit.attempts = 0
		limit.blockedUntil = time.Time{}
	}

	limit.lastAttempt = now

	if success {
		// Reset on successful login
		limit.attempts = 0
		limit.blockedUntil = time.Time{}
	} else {
		limit.attempts++
		if limit.attempts >= rl.maxAttempts {
			limit.blockedUntil = now.Add(rl.blockDuration)
		}
	}
}

// Global rate limiter instance
var globalRateLimiter = NewRateLimiter()
