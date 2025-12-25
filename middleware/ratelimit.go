package middleware

import (
	"Curry2API-go/models"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	// limiterTTL 控制保留每个 IP 限流器的最⻓时间，避免 sync.Map 无限增长
	limiterTTL = 5 * time.Minute
	// cleanupInterval 定期清理过期限流器
	cleanupInterval      = 1 * time.Minute
	defaultRetryAfterSec = 1
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64
}

func newVisitor(limit rate.Limit, burst int) *visitor {
	v := &visitor{
		limiter: rate.NewLimiter(limit, burst),
	}
	v.touch()
	return v
}

func (v *visitor) touch() {
	v.lastSeen.Store(time.Now().UnixNano())
}

func (v *visitor) expired(now time.Time, ttl time.Duration) bool {
	last := v.lastSeen.Load()
	if last == 0 {
		return true
	}
	return now.Sub(time.Unix(0, last)) > ttl
}

type rateLimiterStore struct {
	limit    rate.Limit
	burst    int
	visitors sync.Map
}

func newRateLimiterStore(limit rate.Limit, burst int) *rateLimiterStore {
	store := &rateLimiterStore{
		limit: limit,
		burst: burst,
	}
	go store.cleanupLoop()
	return store
}

func (s *rateLimiterStore) getLimiter(ip string) *rate.Limiter {
	if value, ok := s.visitors.Load(ip); ok {
		v := value.(*visitor)
		v.touch()
		return v.limiter
	}

	v := newVisitor(s.limit, s.burst)
	actual, loaded := s.visitors.LoadOrStore(ip, v)
	if loaded {
		existing := actual.(*visitor)
		existing.touch()
		return existing.limiter
	}
	return v.limiter
}

func (s *rateLimiterStore) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	for now := range ticker.C {
		s.visitors.Range(func(key, value any) bool {
			v := value.(*visitor)
			if v.expired(now, limiterTTL) {
				s.visitors.Delete(key)
			}
			return true
		})
	}
}

// RateLimit 基于 IP 的限流中间件，使用令牌桶算法保护 API
func RateLimit(rps, burst int) gin.HandlerFunc {
	if rps <= 0 {
		rps = 1
	}
	if burst <= 0 {
		burst = 1
	}

	store := newRateLimiterStore(rate.Limit(rps), burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := store.getLimiter(ip)
		if !limiter.Allow() {
			c.Header("Retry-After", strconv.Itoa(defaultRetryAfterSec))
			errorResponse := models.NewErrorResponse(
				"请求过于频繁，请稍后重试",
				"rate_limit_exceeded",
				"rate_limited",
			)
			c.JSON(http.StatusTooManyRequests, errorResponse)
			c.Abort()
			return
		}
		c.Next()
	}
}
