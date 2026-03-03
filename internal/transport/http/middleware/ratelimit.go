package middleware

import (
	"aurora/internal/ratelimit"
	"aurora/internal/transport/http/response"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit applies a Redis-backed token bucket limiter per IP.
// keyPrefix allows isolating limits per route/group (e.g., "auth_login").
func RateLimit(bucket *ratelimit.Bucket, keyPrefix string, capacity, refill int64, period time.Duration) gin.HandlerFunc {
	rate := ratelimit.Rate{
		Capacity: capacity,
		Refill:   refill,
		Period:   period,
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := ratelimit.Key(keyPrefix, ratelimit.ScopeIP, ip)

		res, err := bucket.Allow(c.Request.Context(), key, rate, 1)
		if err != nil {
			response.RespondInternalError(c, "internal server error")
			c.Abort()
			return
		}

		for k, v := range ratelimit.RateLimitHeaders(res) {
			c.Header(k, v)
		}

		if !res.Allowed {
			response.RespondTooManyRequests(c, "rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}
