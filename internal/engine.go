package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nikhil478/rate-limiter/internal/config"
	"github.com/nikhil478/rate-limiter/internal/engine"
	"github.com/nikhil478/rate-limiter/internal/models"
	"github.com/nikhil478/rate-limiter/internal/redis_client"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func StartEngine() {
	ctx := context.Background()
	redisClient := redis_client.CreateRedisInstance()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	cfg, err := loadRateLimiterConfig(cli)
	if err != nil {
		log.Fatal(err)
	}

	if err := buildRateLimiter(cfg, redisClient); err != nil {
		log.Fatal(err)
	}

	go watchRateLimiterConfig(cli, redisClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		limiter := config.GetLimiter()
		if limiter == nil {
			http.Error(w, "rate limiter not initialized", http.StatusInternalServerError)
			return
		}

		isAllowed, headers, err := limiter.Allow(ctx, "abcd")
		if err != nil {
			log.Printf("rate limiter error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		setRateLimitHeaders(w, headers)
		if !isAllowed {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello World\n"))
	})

	serverAddr := ":8080"
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("error while starting server: %v", err)
	}
}

func watchRateLimiterConfig(cli *clientv3.Client, redisClient *redis.Client) {
	watchChan := cli.Watch(context.Background(), "rate_limiter_config")
	for wresp := range watchChan {
		for _, ev := range wresp.Events {
			log.Printf("Config changed: %s : %s\n", ev.Type, ev.Kv.Value)

			var newCfg models.RateLimiterConfig
			if err := json.Unmarshal(ev.Kv.Value, &newCfg); err != nil {
				log.Println("failed to parse updated config:", err)
				continue
			}

			if err := buildRateLimiter(&newCfg, redisClient); err != nil {
				log.Println("failed to rebuild limiter:", err)
				continue
			}

			log.Println("Rate limiter updated dynamically")
		}
	}
}

func buildRateLimiter(cfg *models.RateLimiterConfig, redisClient *redis.Client) error {
	windowSize, err := time.ParseDuration(cfg.WindowSize)
	if err != nil {
		return fmt.Errorf("invalid window_size: %w", err)
	}

	refillRate, err := time.ParseDuration(cfg.RefillRate)
	if err != nil {
		return fmt.Errorf("invalid refill_rate: %w", err)
	}

	newLimiter := engine.NewRateLimiter(redisClient, windowSize, cfg.MaxWindow, cfg.BucketSize, refillRate)
	config.SetLimiter(newLimiter)

	return nil
}

func loadRateLimiterConfig(cli *clientv3.Client) (*models.RateLimiterConfig, error) {
	resp, err := cli.Get(context.Background(), "rate_limiter_config")
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("rate_limiter_config not found in etcd")
	}

	var cfg models.RateLimiterConfig
	if err := json.Unmarshal(resp.Kvs[0].Value, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setRateLimitHeaders(w http.ResponseWriter, headers models.RateLimitResponseHeaders) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(headers.XRateLimitLimit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(headers.XRateLimitRemaining))
	w.Header().Set("X-RateLimit-Reset", strconv.Itoa(headers.XRateLimitReset))
	w.Header().Set("X-RateLimit-Window", strconv.Itoa(headers.XRateLimitWindow))
	w.Header().Set("X-RateLimit-Bucket", strconv.Itoa(headers.XRateLimitBucket))

	if headers.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(headers.RetryAfter))
	}
}
