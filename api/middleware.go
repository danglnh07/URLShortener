package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Middleware for CORS
func (server *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("http://%s:%s", server.config.Domain, server.config.Port))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

		next.ServeHTTP(w, r)
	})
}

// Rate limiter struct, used Token Bucket strategy
type RateLimiter struct {
	tokens     int
	maxToken   int
	refillRate time.Duration
	lastRefill time.Time
	mutex      sync.Mutex
}

// Constructor method for RateLimiter
func NewRateLimiter(maxToken int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxToken,
		maxToken:   maxToken,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Method to check if the current request can pass on, by checking the available token
// while refill token if needed
func (limiter *RateLimiter) Allow() bool {
	// Use mutex to avoid race condition
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	// Refill token
	elapsed := time.Since(limiter.lastRefill)
	refill := int(elapsed / limiter.refillRate)
	if refill > 0 {
		limiter.tokens += refill
		// If tokens exceed max token, we flatten it down
		if limiter.tokens > limiter.maxToken {
			limiter.tokens = limiter.maxToken
		}
		limiter.lastRefill = time.Now()
	}

	// Consume token
	if limiter.tokens > 0 {
		limiter.tokens--
		return true
	}

	// If no token available, simply refuse
	return false
}

// Rate limiting middleware
func (server *Server) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !server.limiter.Allow() {
			server.WriteError(w, http.StatusTooManyRequests, "Too many request at a time")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Chaining middleware, to avoid duplicate code
func (server *Server) ChainingMiddleware(next http.Handler) http.Handler {
	return server.CORSMiddleware(server.RateLimitMiddleware(next))
}
