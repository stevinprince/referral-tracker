package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter provides per-key rate limiting using in-memory token buckets.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a rate limiter that allows `rps` requests per second
// with a burst capacity of `burst`.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
		burst:    burst,
	}
}

func (rate_limiter *RateLimiter) getLimiter(key string) *rate.Limiter {
	rate_limiter.mu.Lock()
	defer rate_limiter.mu.Unlock()

	if limiter, exists := rate_limiter.limiters[key]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate_limiter.rate, rate_limiter.burst)
	rate_limiter.limiters[key] = limiter
	return limiter
}

// ByIP returns a middleware that rate-limits by client IP.
func (rl *RateLimiter) ByIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !rl.getLimiter(key).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "Too many requests. Please try again later.",
				},
			})
			return
		}
		c.Next()
	}
}

// BySession returns a middleware that rate-limits by session cookie.
func (rl *RateLimiter) BySession() gin.HandlerFunc {
	return func(c *gin.Context) {
		key, err := c.Cookie(session_cookie_name)
		if err != nil || key == "" {
			// No session — let the auth middleware handle rejection
			c.Next()
			return
		}

		if !rl.getLimiter(key).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "Too many requests. Please try again later.",
				},
			})
			return
		}
		c.Next()
	}
}
