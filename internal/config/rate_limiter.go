package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nikhil478/rate-limiter/internal/engine"
	"github.com/nikhil478/rate-limiter/internal/models"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
)


var (
	currentLimiter *engine.RateLimiter
	limiterMu      sync.RWMutex
)

func GetLimiter() *engine.RateLimiter {
	limiterMu.RLock()
	defer limiterMu.RUnlock()
	return currentLimiter
}

func SetLimiter(newLimiter *engine.RateLimiter) {
	limiterMu.Lock()
	defer limiterMu.Unlock()
	currentLimiter = newLimiter
}


func InitRateLimiter(redisClient *redis.Client) (*engine.RateLimiter, error) {
    cfg, err := LoadRateLimiterConfig()
    if err != nil {
        return nil, err
    }
	
    windowSize, err := time.ParseDuration(cfg.WindowSize)
    if err != nil {
        return nil, fmt.Errorf("invalid window_size: %w", err)
    }

    refillRate, err := time.ParseDuration(cfg.RefillRate)
    if err != nil {
        return nil, fmt.Errorf("invalid refill_rate: %w", err)
    }

    limiter := engine.NewRateLimiter(
        redisClient,
        windowSize,
        cfg.MaxWindow,
        cfg.BucketSize,
        refillRate,
    )

    return limiter, nil
}



func LoadRateLimiterConfig() (*models.RateLimiterConfig, error) {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, "rate_limiter_config")
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("rate_limiter_config not found in etcd")
	}

	var cfg models.RateLimiterConfig
	err = json.Unmarshal(resp.Kvs[0].Value, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
