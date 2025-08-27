package models


type RateLimiterConfig struct {
    WindowSize string `json:"window_size"`
    MaxWindow  int    `json:"max_window"`
    BucketSize int    `json:"bucket_size"`
    RefillRate string `json:"refill_rate"`
}