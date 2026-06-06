package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
	"github.com/gin-gonic/gin"
)

// IPRateLimiter rate limiter per IP
type IPRateLimiter struct {
	ips     map[string]*rate.Limiter
	mu      sync.RWMutex
	rate    rate.Limit
	burst   int
}

// NewIPRateLimiter creates a new rate limiter
// r = requests per second, b = burst size
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:   make(map[string]*rate.Limiter),
		rate:  r,
		burst: b,
	}
}

// GetLimiter returns the rate limiter for the given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.ips[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware returns a Gin middleware for rate limiting
func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := limiter.GetLimiter(ip)
		if !l.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
