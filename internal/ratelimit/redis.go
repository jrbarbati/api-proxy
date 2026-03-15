package ratelimit

import (
	"api-proxy/internal/model"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client    *redis.Client
	rw        sync.RWMutex
	orgLimits map[int]int
	saLimits  map[int]int
}

func NewRedisRateLimiter(url string) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: redis.NewClient(&redis.Options{
			Addr: url,
		}),
		rw:        sync.RWMutex{},
		orgLimits: make(map[int]int),
		saLimits:  make(map[int]int),
	}
}

func (rrl *RedisRateLimiter) AllowRequest(orgID, saID int) bool {
	ctx := context.Background()

	rrl.rw.RLock()
	orgLimit, orgHasLimit := rrl.orgLimits[orgID]
	saLimit, saHasLimit := rrl.saLimits[saID]
	rrl.rw.RUnlock()

	if !orgHasLimit && !saHasLimit {
		return true
	}

	if orgHasLimit {
		if !rrl.allow(ctx, fmt.Sprintf("org:%d", orgID), orgLimit) {
			return false
		}
	}

	if saHasLimit {
		if !rrl.allow(ctx, fmt.Sprintf("sa:%d", saID), saLimit) {
			return false
		}
	}

	return true
}

func (rrl *RedisRateLimiter) StartSync(ctx context.Context, interval time.Duration, findRateLimits func() ([]*model.RateLimit, error)) {
	rrl.syncCache(findRateLimits)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rrl.syncCache(findRateLimits)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (rrl *RedisRateLimiter) allow(ctx context.Context, bucketKey string, limit int) bool {
	// TODO: Implement
	// 	Ask redis for the current token count (remaining?) based on the bucket key
	// 	If it has one, reduce the number by 1 and allow the request.
	return false
}

func (rrl *RedisRateLimiter) syncCache(findRateLimits func() ([]*model.RateLimit, error)) {
	slog.Info("started rate limit cache sync...")

	limits, err := findRateLimits()

	if err != nil {
		slog.Error("error finding rate limits", "err", err)
		return
	}

	orgLimits := make(map[int]struct{})
	saLimits := make(map[int]struct{})

	rrl.rw.Lock()
	defer rrl.rw.Unlock()

	for _, limit := range limits {
		if limit.ServiceAccountID == nil {
			orgLimits[limit.OrgID] = struct{}{}

			rrl.orgLimits[limit.OrgID] = limit.LimitPerMinute
		} else {
			saLimits[*limit.ServiceAccountID] = struct{}{}

			rrl.saLimits[*limit.ServiceAccountID] = limit.LimitPerMinute
		}
	}

	for k := range rrl.orgLimits {
		if _, ok := orgLimits[k]; !ok {
			delete(rrl.orgLimits, k)
		}
	}

	for k := range rrl.saLimits {
		if _, ok := saLimits[k]; !ok {
			delete(rrl.saLimits, k)
		}
	}

	slog.Info("finished rate limit cache sync...")
}
