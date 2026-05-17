package ratelimit

import (
    "context"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
    rdb       *redis.Client
    rate      int
    window    time.Duration
    keyPrefix string
}

func NewRateLimiter(rdb *redis.Client, rate int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        rdb:       rdb,
        rate:      rate,
        window:    window,
        keyPrefix: "rl:",
    }
}

func (rl *RateLimiter) Allow(userID string) (bool, error) {
    key := rl.keyPrefix + userID
    now := time.Now().UnixNano()
    windowStart := now - rl.window.Nanoseconds()

    script := `
        local key = KEYS[1]
        local now = tonumber(ARGV[1])
        local windowStart = tonumber(ARGV[2])
        local rate = tonumber(ARGV[3])
        redis.call('ZREMRANGEBYSCORE', key, '-inf', windowStart)
        local current = redis.call('ZCARD', key)
        if current < rate then
            redis.call('ZADD', key, now, now)
            redis.call('EXPIRE', key, ARGV[4])
            return 1
        else
            return 0
        end
    `
    resp, err := rl.rdb.Eval(context.Background(), script,
        []string{key},
        now, windowStart, rl.rate, int(rl.window.Seconds())).Int()
    if err != nil {
        return false, err
    }
    return resp == 1, nil
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("user_id")
        if !exists {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        allowed, err := limiter.Allow(userID.(string))
        if err != nil {
            c.AbortWithStatusJSON(500, gin.H{"error": "rate limit internal error"})
            return
        }
        if !allowed {
            c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
            return
        }
        c.Next()
    }
}
