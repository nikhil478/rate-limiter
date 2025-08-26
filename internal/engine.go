package internal

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nikhil478/rate-limiter/internal/engine"
	"github.com/nikhil478/rate-limiter/internal/models"
	"github.com/nikhil478/rate-limiter/internal/redis_client"
)

func StartEngine() {

	ctx := context.Background()
	redisClient := redis_client.CreateRedisInstance()
	rateLimiter := engine.NewRateLimiter(redisClient, time.Minute, 10, 3, 6*time.Second)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		isAllowed, headers, err := rateLimiter.Allow(ctx, "abcd")

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
