// Package middleware also holds request-throttling logic to prevent abuse
// (e.g. brute-forcing login credentials).
package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// visitor tracks the rate limiter for a single IP address.
type visitor struct {
	limiter *rate.Limiter
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex // protects the visitors map from concurrent access
)

// getLimiter returns the rate limiter for a given IP, creating one if this
// is the first time we've seen that IP.
func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		// Allow 5 requests, refilling at 1 every 2 seconds — deliberately
		// strict, since this is meant only for sensitive auth endpoints.
		limiter := rate.NewLimiter(0.5, 5)
		visitors[ip] = &visitor{limiter: limiter}
		return limiter
	}
	return v.limiter
}

// RateLimit is middleware that rejects requests once an IP exceeds its
// allowed rate. Intended for /login and /signup specifically, not the
// whole app.
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests — please slow down"})
			c.Abort()
			return
		}

		c.Next()
	}
}
