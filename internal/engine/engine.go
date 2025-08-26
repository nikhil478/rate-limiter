package engine

import (
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

