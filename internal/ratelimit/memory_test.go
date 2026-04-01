package ratelimit

import (
	"context"
	"sync"
	"testing"
)

func TestMemoryRateLimiter_AllowRequest(t *testing.T) {
	scenarios := []struct {
		name        string
		orgID       int
		saID        int
		rateLimiter MemoryRateLimiter
		expected    bool
	}{
		{
			name:  "No Rate Limits",
			orgID: 1,
			saID:  1,
			rateLimiter: MemoryRateLimiter{
				rw:        sync.RWMutex{},
				orgLimits: map[int]tokenBucket{},
				saLimits:  map[int]tokenBucket{},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: true,
		},
		{
			name:  "Org Limit Allowed",
			orgID: 1,
			saID:  1,
			rateLimiter: MemoryRateLimiter{
				rw: sync.RWMutex{},
				orgLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: true},
				},
				saLimits: map[int]tokenBucket{},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: true,
		},
		{
			name:  "Org Limit Not Allowed",
			orgID: 1,
			saID:  1,
			rateLimiter: MemoryRateLimiter{
				rw: sync.RWMutex{},
				orgLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: false},
				},
				saLimits: map[int]tokenBucket{},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: false,
		},
		{
			name:  "SA Limit Allowed",
			orgID: 1,
			saID:  1,
			rateLimiter: MemoryRateLimiter{
				rw:        sync.RWMutex{},
				orgLimits: map[int]tokenBucket{},
				saLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: true},
				},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: true,
		},
		{
			name:  "SA Limit Not Allowed",
			orgID: 1,
			saID:  1,
			rateLimiter: MemoryRateLimiter{
				rw:        sync.RWMutex{},
				orgLimits: map[int]tokenBucket{},
				saLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: false},
				},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: false,
		},
		{
			name:  "Both Limit Allowed",
			orgID: 1,
			saID:  1,
			rateLimiter: MemoryRateLimiter{
				rw: sync.RWMutex{},
				orgLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: true},
				},
				saLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: true},
				},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: true,
		},
		{
			name:  "Both Limit Not Allowed (Org)",
			orgID: 1,
			saID:  1,
			rateLimiter: MemoryRateLimiter{
				rw: sync.RWMutex{},
				orgLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: false},
				},
				saLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: true},
				},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{
						capacity:   limit,
						allowToken: false,
					}
				},
			},
			expected: false,
		},
		{
			name:  "Both Limit Not Allowed (SA)",
			orgID: 1,
			saID:  1,
			rateLimiter: MemoryRateLimiter{
				rw: sync.RWMutex{},
				orgLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: true},
				},
				saLimits: map[int]tokenBucket{
					1: &mockTokenBucket{allowToken: false},
				},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			actual := scenario.rateLimiter.AllowRequest(scenario.orgID, scenario.saID)

			if scenario.expected != actual {
				t.Fatalf("expected: %v, actual: %v", scenario.expected, actual)
			}
		})
	}
}

func TestMemoryRateLimiter_StartSync(t *testing.T) {

}

func TestMemoryRateLimiter_orgBucket(t *testing.T) {
	scenarios := []struct {
		name        string
		orgID       int
		rateLimiter MemoryRateLimiter
		expected    bool
	}{
		{
			name:  "No Bucket",
			orgID: 5,
			rateLimiter: MemoryRateLimiter{
				rw: sync.RWMutex{},
				orgLimits: map[int]tokenBucket{
					1: &mockTokenBucket{capacity: 1},
					2: &mockTokenBucket{capacity: 2},
					3: &mockTokenBucket{capacity: 3},
				},
				saLimits: map[int]tokenBucket{
					11: &mockTokenBucket{capacity: 11},
					12: &mockTokenBucket{capacity: 12},
					13: &mockTokenBucket{capacity: 13}},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: false,
		},
		{
			name:  "Has Bucket",
			orgID: 2,
			rateLimiter: MemoryRateLimiter{
				rw: sync.RWMutex{},
				orgLimits: map[int]tokenBucket{
					1: &mockTokenBucket{capacity: 1},
					2: &mockTokenBucket{capacity: 2},
					3: &mockTokenBucket{capacity: 3},
				},
				saLimits: map[int]tokenBucket{
					11: &mockTokenBucket{capacity: 11},
					12: &mockTokenBucket{capacity: 12},
					13: &mockTokenBucket{capacity: 13}},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			orgBucket, actual := scenario.rateLimiter.orgBucket(scenario.orgID)

			if scenario.expected != actual {
				t.Fatalf("expected: %v, actual: %v", scenario.expected, actual)
			}

			if scenario.expected && orgBucket == nil {
				t.Fatalf("orgBucket is nil, expected non-nil")
			}

			if !scenario.expected && orgBucket != nil {
				t.Fatalf("orgBucket is non-nil, expected nil")
			}
		})
	}
}

func TestMemoryRateLimiter_saBucket(t *testing.T) {
	scenarios := []struct {
		name        string
		saID        int
		rateLimiter MemoryRateLimiter
		expected    bool
	}{
		{
			name: "No Bucket",
			saID: 15,
			rateLimiter: MemoryRateLimiter{
				rw: sync.RWMutex{},
				orgLimits: map[int]tokenBucket{
					1: &mockTokenBucket{capacity: 1},
					2: &mockTokenBucket{capacity: 2},
					3: &mockTokenBucket{capacity: 3},
				},
				saLimits: map[int]tokenBucket{
					11: &mockTokenBucket{capacity: 11},
					12: &mockTokenBucket{capacity: 12},
					13: &mockTokenBucket{capacity: 13}},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: false,
		},
		{
			name: "Has Bucket",
			saID: 12,
			rateLimiter: MemoryRateLimiter{
				rw: sync.RWMutex{},
				orgLimits: map[int]tokenBucket{
					1: &mockTokenBucket{capacity: 1},
					2: &mockTokenBucket{capacity: 2},
					3: &mockTokenBucket{capacity: 3},
				},
				saLimits: map[int]tokenBucket{
					11: &mockTokenBucket{capacity: 11},
					12: &mockTokenBucket{capacity: 12},
					13: &mockTokenBucket{capacity: 13}},
				newTokenBucket: func(limit int) tokenBucket {
					return &mockTokenBucket{capacity: limit}
				},
			},
			expected: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			saBucket, actual := scenario.rateLimiter.saBucket(scenario.saID)

			if scenario.expected != actual {
				t.Fatalf("expected: %v, actual: %v", scenario.expected, actual)
			}

			if scenario.expected && saBucket == nil {
				t.Fatalf("saBucket is nil, expected non-nil")
			}

			if !scenario.expected && saBucket != nil {
				t.Fatalf("saBucket is non-nil, expected nil")
			}
		})
	}
}

type mockTokenBucket struct {
	capacity       int
	allowToken     bool
	startCallCount int
	stopCallCount  int
}

func (mtb *mockTokenBucket) Start(context.Context) {
	mtb.startCallCount++
}

func (mtb *mockTokenBucket) Stop() {
	mtb.stopCallCount++
}

func (mtb *mockTokenBucket) UpdateCapacity(newCapacity int) {
	mtb.capacity = newCapacity
}

func (mtb *mockTokenBucket) requestToken() bool {
	return mtb.allowToken
}
