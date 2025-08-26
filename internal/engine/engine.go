package engine

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type RateLimiter struct {
	RedisClient *redis.Client
	WindowSize  time.Duration
	MaxWindow   int
	BucketSize  int
	RefillRate  time.Duration
}

/*
NewRateLimiter creates and returns an instance of a custom rate limiter.

@redisClient - Interface to the Redis client. This function expects a valid Redis client instance.
@windowSize  - Duration of the sliding window.
@maxWindow   - Maximum number of requests allowed within the window.
@bucketSize  - Capacity of the token bucket.
@refillRate  - Interval at which tokens are refilled into the bucket.

This custom rate limiter combines the Token Bucket and Sliding Window algorithms. The token bucket allows controlled request refilling, while
the sliding window ensures fair request distribution within a time window. The trade-off is higher operational cost due to the added complexity.
*/
func NewRateLimiter(redisClient *redis.Client, windowSize time.Duration, maxWindow int, bucketSize int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		RedisClient: redisClient,
		WindowSize:  windowSize,
		MaxWindow:   maxWindow,
		BucketSize:  bucketSize,
		RefillRate:  refillRate,
	}
}

/*
Allow checks whether a given key (user/client) is allowed to make a request
based on a hybrid rate limiting strategy (Sliding Window + Token Bucket).

Steps:
1. Sliding Window check:
  - Use a Redis sorted set (logKey) to track request timestamps.
  - Remove old entries outside the current window.
  - Count requests inside the active window.
  - If requests exceed maxWindow, reject immediately.

2. Token Bucket check:
  - Retrieve the current token count (tokens) and last refill timestamp (lastRefill).
  - Refill tokens based on elapsed time since lastRefill and configured refillRate.
  - Ensure tokens do not exceed bucketSize.
  - If no tokens are available, reject.
  - Otherwise, consume a token and allow the request.

This hybrid model ensures fairness (Sliding Window) while preventing bursts (Token Bucket).
*/
func (rl *RateLimiter) Allow(key string) (bool, error) {

	now := time.Now().Unix()
	logKey := fmt.Sprintf("log:%s", key)

	minScore := "0"
	maxScore := fmt.Sprintf("%d", now-int64(rl.WindowSize.Seconds()))
	if err := rl.RedisClient.ZRemRangeByScore(logKey, minScore, maxScore).Err(); err != nil {
		return false, err
	}

	count, err := rl.RedisClient.ZCard(logKey).Result()
	if err != nil {
		return false, err
	}

	if count >= int64(rl.MaxWindow) {
		return false, nil
	}

	if err := rl.RedisClient.ZAdd(logKey, redis.Z{
		Score:  float64(now),
		Member: now,
	}).Err(); err != nil {
		return false, err
	}

	if err := rl.RedisClient.Expire(logKey, rl.WindowSize).Err(); err != nil {
		return false, err
	}

	tokenKey := fmt.Sprintf("tokens:%s", key)
	timeKey := fmt.Sprintf("lastRefill:%s", key)

	tokens, err := rl.RedisClient.Get(tokenKey).Int()
	if err == redis.Nil {
		tokens = rl.BucketSize
	} else if err != nil {
		return false, err
	}

	lastRefill, err := rl.RedisClient.Get(timeKey).Int64()
	if err == redis.Nil {
		lastRefill = now
	} else if err != nil {
		return false, err
	}

	elapsed := now - lastRefill
	refillTokens := int(elapsed) / int(rl.RefillRate.Seconds())
	if refillTokens > 0 {
		tokens = min(tokens+refillTokens, rl.BucketSize)
		lastRefill = now
	}

	if tokens <= 0 {
		return false, nil
	}
	tokens--

	if err := rl.RedisClient.Set(tokenKey, tokens, 0).Err(); err != nil {
		return false, err
	}
	if err := rl.RedisClient.Set(timeKey, lastRefill, 0).Err(); err != nil {
		return false, err
	}

	return true, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
