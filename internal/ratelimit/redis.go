package ratelimit

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(url string) *RateLimiter {
	return &RateLimiter{
		url: url,
	}
}

func (rl *RateLimiter) AllowRequest(orgID, saID int) bool {
	return false
}
