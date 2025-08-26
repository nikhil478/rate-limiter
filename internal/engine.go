package internal

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/nikhil478/rate-limiter/internal/engine"
	"github.com/nikhil478/rate-limiter/internal/redis_client"
)

func StartEngine() {

	ctx := context.Background()
	redisClient := redis_client.CreateRedisInstance()
	rateLimiter := engine.NewRateLimiter(redisClient, time.Minute, 10, 3, 6*time.Second)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		isAllowed, err := rateLimiter.Allow(ctx, "abcd")

		if err != nil {
			log.Printf("rate limiter error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

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
