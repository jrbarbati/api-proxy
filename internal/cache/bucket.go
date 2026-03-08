package cache

import (
	"context"
	"time"
)

type Bucket struct {
	capacity       int
	tokens         int
	lastRefill     time.Time
	tokenRequested chan tokenRequest
	capacityUpdate chan int
	stop           context.CancelFunc
}

type tokenRequest struct {
	response chan bool
}

func NewBucket(capacity int) *Bucket {
	return &Bucket{
		capacity:       capacity,
		tokens:         capacity,
		lastRefill:     time.Now(),
		tokenRequested: make(chan tokenRequest),
		capacityUpdate: make(chan int),
	}
}

func newTokenRequest() tokenRequest {
	return tokenRequest{
		response: make(chan bool),
	}
}

func (b *Bucket) RequestToken() bool {
	request := newTokenRequest()
	b.tokenRequested <- request
	return <-request.response
}

func (b *Bucket) UpdateCapacity(newCapacity int) {
	b.capacityUpdate <- newCapacity
}

func (b *Bucket) Start(ctx context.Context) {
	bucketCtx, cancel := context.WithCancel(ctx)
	b.stop = cancel

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case newCapacity := <-b.capacityUpdate:
				b.updateCapacity(newCapacity)
			case tokenReq := <-b.tokenRequested:
				tokenReq.response <- b.takeToken()
			case <-ticker.C:
				b.refill()
			case <-bucketCtx.Done():
				return
			}
		}
	}()
}

func (b *Bucket) Stop() {
	if b.stop != nil {
		b.stop()
	}
}

func (b *Bucket) takeToken() bool {
	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

func (b *Bucket) updateCapacity(newCapacity int) {
	difference := b.capacity - newCapacity

	if difference == 0 {
		return
	}

	b.capacity = newCapacity
	b.tokens = max(0, b.tokens-difference)
}

func (b *Bucket) refill() {
	b.tokens = b.capacity
	b.lastRefill = time.Now()
}
