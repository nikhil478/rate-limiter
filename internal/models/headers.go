package models


type RateLimitResponseHeaders struct {
	XRateLimitLimit     int 
	XRateLimitRemaining int
	XRateLimitReset     int
	RetryAfter          int
	XRateLimitWindow    int
	XRateLimitBucket    int
}
