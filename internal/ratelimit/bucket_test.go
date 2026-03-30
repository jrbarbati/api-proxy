package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestBucket_requestToken(t *testing.T) {
	scenarios := []struct {
		name          string
		bucket        *bucket
		startBucket   bool
		expectAllowed bool
	}{
		{
			name: "no available tokens",
			bucket: &bucket{
				capacity:       100,
				tokens:         0,
				tokenRequested: make(chan tokenRequest),
			},
			startBucket:   true,
			expectAllowed: false,
		},
		{
			name: "available tokens",
			bucket: &bucket{
				capacity:       100,
				tokens:         50,
				tokenRequested: make(chan tokenRequest),
			},
			startBucket:   true,
			expectAllowed: true,
		},
		{
			name: "unstarted bucket",
			bucket: &bucket{
				capacity:       100,
				tokens:         50,
				tokenRequested: make(chan tokenRequest),
			},
			startBucket:   false,
			expectAllowed: false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			if scenario.startBucket {
				scenario.bucket.Start(t.Context())
			}

			time.Sleep(5 * time.Millisecond)

			allowed := scenario.bucket.requestToken()

			time.Sleep(5 * time.Millisecond)

			if scenario.expectAllowed != allowed {
				t.Fatalf("expected %v, got %v", scenario.expectAllowed, allowed)
			}
		})
	}
}

func TestBucket_Stop(t *testing.T) {
	t.Run("non-nil", func(t *testing.T) {
		called := false
		b := &bucket{
			stop: func() {
				called = true
			},
		}

		b.Stop()

		if !called {
			t.Fatal("expected stop to be called")
		}
	})

	t.Run("nil", func(t *testing.T) {
		called := false
		b := &bucket{
			stop: nil,
		}

		b.Stop()

		if called {
			t.Fatal("expected stop to be called")
		}
	})
}

func TestBucket_takeToken(t *testing.T) {
	scenarios := []struct {
		name            string
		bucket          *bucket
		expectedAllowed bool
		expectedTokens  int
	}{
		{
			name: "no tokens",
			bucket: &bucket{
				capacity: 100,
				tokens:   0,
			},
			expectedAllowed: false,
		},
		{
			name: "has tokens",
			bucket: &bucket{
				capacity: 100,
				tokens:   52,
			},
			expectedAllowed: true,
			expectedTokens:  51,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			allowed := scenario.bucket.takeToken()

			if scenario.expectedAllowed != allowed {
				t.Fatalf("expected allowed %v, got %v", scenario.expectedAllowed, allowed)
			}

			if scenario.expectedTokens != scenario.bucket.tokens {
				t.Fatalf("expected tokens %v, got %v", scenario.expectedTokens, scenario.bucket.tokens)
			}
		})
	}
}

func TestBucket_updateCapacity(t *testing.T) {
	scenarios := []struct {
		name             string
		bucket           *bucket
		updateCapacity   int
		expectedCapacity int
		expectedTokens   int
	}{
		{
			name: "positive update",
			bucket: &bucket{
				capacity: 100,
				tokens:   65,
			},
			updateCapacity:   150,
			expectedCapacity: 150,
			expectedTokens:   115,
		},
		{
			name: "negative update",
			bucket: &bucket{
				capacity: 100,
				tokens:   65,
			},
			updateCapacity:   50,
			expectedCapacity: 50,
			expectedTokens:   15,
		},
		{
			name: "negative update, no more tokens",
			bucket: &bucket{
				capacity: 100,
				tokens:   65,
			},
			updateCapacity:   10,
			expectedCapacity: 10,
			expectedTokens:   0,
		},
		{
			name: "no-op update",
			bucket: &bucket{
				capacity: 100,
				tokens:   65,
			},
			updateCapacity:   100,
			expectedCapacity: 100,
			expectedTokens:   65,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			scenario.bucket.updateCapacity(scenario.updateCapacity)

			if scenario.expectedCapacity != scenario.bucket.capacity {
				t.Fatalf("expected capacity %v, got %v", scenario.expectedCapacity, scenario.bucket.capacity)
			}

			if scenario.expectedTokens != scenario.bucket.tokens {
				t.Fatalf("expected tokens %v, got %v", scenario.expectedCapacity, scenario.bucket.capacity)
			}
		})
	}
}

func TestBucket_refill(t *testing.T) {
	scenarios := []struct {
		name               string
		bucket             *bucket
		expectedTokens     int
		expectedRefillTime time.Time
		timeDiffAllowed    time.Duration
	}{
		{
			name: "refill",
			bucket: &bucket{
				capacity: 100,
				tokens:   52,
			},
			expectedTokens:     100,
			expectedRefillTime: time.Now(),
			timeDiffAllowed:    1 * time.Second,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			scenario.bucket.refill()

			if scenario.expectedTokens != scenario.bucket.tokens {
				t.Fatalf("expected tokens %v, got %v", scenario.expectedTokens, scenario.bucket.tokens)
			}

			if scenario.bucket.lastRefill.Sub(scenario.expectedRefillTime) > scenario.timeDiffAllowed {
				t.Fatalf("difference between expected and actual last refill time greater than allowed threshold")
			}
		})
	}
}

type callCounter struct {
	mux     sync.Mutex
	counter int
}

func newCallCounter() *callCounter {
	return &callCounter{
		mux:     sync.Mutex{},
		counter: 0,
	}
}
