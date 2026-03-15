package ratelimit

import (
	"api-proxy/internal/model"
	"context"
	"log/slog"
	"sync"
	"time"
)

type RateLimitCache struct {
	rw        sync.RWMutex
	orgLimits map[int]*Bucket
	saLimits  map[int]*Bucket
}

func NewRateLimitCache() *RateLimitCache {
	return &RateLimitCache{
		rw:        sync.RWMutex{},
		orgLimits: make(map[int]*Bucket),
		saLimits:  make(map[int]*Bucket),
	}
}

func (rlc *RateLimitCache) AllowRequest(orgID, saID int) bool {
	orgBucket, orgHasBucket := rlc.orgBucket(orgID)
	saBucket, saHasBucket := rlc.saBucket(saID)

	if !orgHasBucket && !saHasBucket {
		return true
	}

	return (orgBucket == nil || orgBucket.RequestToken()) && (saBucket == nil || saBucket.RequestToken())
}

func (rlc *RateLimitCache) StartSync(ctx context.Context, interval time.Duration, findRateLimits func() ([]*model.RateLimit, error)) {
	rlc.syncCache(ctx, findRateLimits)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rlc.syncCache(ctx, findRateLimits)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (rlc *RateLimitCache) orgBucket(orgID int) (*Bucket, bool) {
	rlc.rw.RLock()
	defer rlc.rw.RUnlock()

	bucket, ok := rlc.orgLimits[orgID]
	return bucket, ok
}

func (rlc *RateLimitCache) saBucket(saID int) (*Bucket, bool) {
	rlc.rw.RLock()
	defer rlc.rw.RUnlock()

	bucket, ok := rlc.saLimits[saID]
	return bucket, ok
}

func (rlc *RateLimitCache) syncCache(ctx context.Context, findRateLimits func() ([]*model.RateLimit, error)) {
	slog.Info("started rate limit cache sync...")

	limits, err := findRateLimits()

	if err != nil {
		slog.Error("error finding rate limits", "err", err)
		return
	}

	orgLimitsMap := make(map[int]struct{})
	saLimitsMap := make(map[int]struct{})

	rlc.rw.Lock()
	defer rlc.rw.Unlock()

	for _, limit := range limits {
		if limit.ServiceAccountID == nil {
			orgLimitsMap[limit.OrgID] = struct{}{}

			if _, ok := rlc.orgLimits[limit.OrgID]; !ok {
				rlc.orgLimits[limit.OrgID] = NewBucket(limit.LimitPerMinute)
				rlc.orgLimits[limit.OrgID].Start(ctx)
			} else {
				rlc.orgLimits[limit.OrgID].UpdateCapacity(limit.LimitPerMinute)
			}
		} else {
			saLimitsMap[*limit.ServiceAccountID] = struct{}{}

			if _, ok := rlc.saLimits[*limit.ServiceAccountID]; !ok {
				rlc.saLimits[*limit.ServiceAccountID] = NewBucket(limit.LimitPerMinute)
				rlc.saLimits[*limit.ServiceAccountID].Start(ctx)
			} else {
				rlc.saLimits[*limit.ServiceAccountID].UpdateCapacity(limit.LimitPerMinute)
			}
		}
	}

	for k, v := range rlc.orgLimits {
		if _, ok := orgLimitsMap[k]; !ok {
			v.Stop()
			delete(rlc.orgLimits, k)
		}
	}

	for k, v := range rlc.saLimits {
		if _, ok := saLimitsMap[k]; !ok {
			v.Stop()
			delete(rlc.saLimits, k)
		}
	}

	slog.Info("finished rate limit cache sync...")
}
