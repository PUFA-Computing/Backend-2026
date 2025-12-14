package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Improved rate limiter with per-IP token buckets and automatic cleanup
type TokenBucket struct {
	tokens         int
	lastRefill     time.Time
	refillInterval time.Duration
	maxTokens      int
	mu             sync.Mutex // Per-bucket mutex instead of global
}

type RateLimiterV2 struct {
	buckets        map[string]*TokenBucket
	mu             sync.RWMutex // Read-write mutex for better concurrency
	maxTokens      int
	refillInterval time.Duration
	cleanupTicker  *time.Ticker
}

func NewRateLimiterV2(maxTokens int, refillInterval time.Duration) *RateLimiterV2 {
	rl := &RateLimiterV2{
		buckets:        make(map[string]*TokenBucket),
		maxTokens:      maxTokens,
		refillInterval: refillInterval,
		cleanupTicker:  time.NewTicker(5 * time.Minute), // Cleanup every 5 minutes
	}
	
	// Start cleanup goroutine to prevent memory leaks
	go rl.cleanup()
	
	return rl
}

func (rl *RateLimiterV2) getBucket(key string) *TokenBucket {
	// Try read lock first (faster for existing buckets)
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()
	
	if exists {
		return bucket
	}
	
	// Need to create new bucket, use write lock
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// Double-check after acquiring write lock
	if bucket, exists := rl.buckets[key]; exists {
		return bucket
	}
	
	// Create new bucket
	bucket = &TokenBucket{
		tokens:         rl.maxTokens,
		lastRefill:     time.Now(),
		refillInterval: rl.refillInterval,
		maxTokens:      rl.maxTokens,
	}
	rl.buckets[key] = bucket
	return bucket
}

func (tb *TokenBucket) allow() (bool, int, time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	// Refill tokens based on elapsed time
	if elapsed >= tb.refillInterval {
		refillAmount := int(elapsed / tb.refillInterval)
		tb.tokens = min(tb.tokens+refillAmount*tb.maxTokens, tb.maxTokens)
		tb.lastRefill = now
	}
	
	nextRefill := tb.lastRefill.Add(tb.refillInterval)
	
	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true, tb.tokens, nextRefill
	}
	
	return false, 0, nextRefill
}

// Cleanup old buckets to prevent memory leaks
func (rl *RateLimiterV2) cleanup() {
	for range rl.cleanupTicker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			bucket.mu.Lock()
			// Remove buckets that haven't been used in 10 minutes
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.buckets, key)
			}
			bucket.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiterV2) Stop() {
	rl.cleanupTicker.Stop()
}

// Per-endpoint rate limiters
var (
	generalLimiter *RateLimiterV2
	authLimiter    *RateLimiterV2
	voteLimiter    *RateLimiterV2
	uploadLimiter  *RateLimiterV2
	publicLimiter  *RateLimiterV2
	once           sync.Once
)

func initLimiters() {
	once.Do(func() {
		generalLimiter = NewRateLimiterV2(100, time.Minute)  // 100 req/min
		authLimiter = NewRateLimiterV2(10, time.Minute)      // 10 req/min
		voteLimiter = NewRateLimiterV2(5, time.Minute)       // 5 req/min
		uploadLimiter = NewRateLimiterV2(3, time.Minute)     // 3 req/min
		publicLimiter = NewRateLimiterV2(200, time.Minute)   // 200 req/min
	})
}

func RateLimiterMiddleware(maxTokens int, refillInterval time.Duration, key string) gin.HandlerFunc {
	initLimiters()
	
	return func(c *gin.Context) {
		// Skip rate limiting for health checks
		if c.Request.URL.Path == "/api/v1/health" {
			c.Next()
			return
		}
		
		ip := c.ClientIP()
		limiterKey := ip + ":" + key
		
		// Select appropriate limiter based on key
		var limiter *RateLimiterV2
		switch key {
		case "auth":
			limiter = authLimiter
		case "vote":
			limiter = voteLimiter
		case "upload":
			limiter = uploadLimiter
		case "public":
			limiter = publicLimiter
		default:
			limiter = generalLimiter
		}
		
		bucket := limiter.getBucket(limiterKey)
		allowed, tokensLeft, nextRefill := bucket.allow()
		
		// Add headers to indicate rate limiting status
		c.Writer.Header().Set("X-Rate-Limit-Limit", strconv.Itoa(limiter.maxTokens))
		c.Writer.Header().Set("X-Rate-Limit-Remaining", strconv.Itoa(tokensLeft))
		c.Writer.Header().Set("X-Rate-Limit-Reset", strconv.FormatInt(nextRefill.Unix(), 10))
		
		if !allowed {
			retryAfter := int(time.Until(nextRefill).Seconds())
			c.Writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": retryAfter,
			})
			return
		}
		
		c.Next()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
