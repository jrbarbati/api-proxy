package middleware

import (
	"context"
	"time"
)

type bucket struct {
	capacity       int
	tokens         int
	lastRefill     time.Time
	tokenRequested chan tokenRequest
}

type tokenRequest struct {
	response chan bool
}

func newBucket(capacity int) *bucket {
	return &bucket{
		capacity:       capacity,
		tokens:         capacity,
		lastRefill:     time.Now(),
		tokenRequested: make(chan tokenRequest),
	}
}

func newTokenRequest() tokenRequest {
	return tokenRequest{
		response: make(chan bool),
	}
}

func (b *bucket) takeToken() bool {
	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

func (b *bucket) refill() {
	b.tokens = b.capacity
	b.lastRefill = time.Now()
}

func (b *bucket) start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case req := <-b.tokenRequested:
				req.response <- b.takeToken()
			case <-ticker.C:
				b.refill()
			case <-ctx.Done():
				return
			}
		}
	}()
}
