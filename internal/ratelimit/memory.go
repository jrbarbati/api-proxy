package ratelimit

import (
	"api-proxy/internal/model"
	"context"
	"log/slog"
	"sync"
	"time"
)

type MemoryRateLimiter struct {
	rw        sync.RWMutex
	orgLimits map[int]*Bucket
	saLimits  map[int]*Bucket
}

func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		rw:        sync.RWMutex{},
		orgLimits: make(map[int]*Bucket),
		saLimits:  make(map[int]*Bucket),
	}
}

func (mrl *MemoryRateLimiter) AllowRequest(orgID, saID int) bool {
	orgBucket, orgHasBucket := mrl.orgBucket(orgID)
	saBucket, saHasBucket := mrl.saBucket(saID)

	if !orgHasBucket && !saHasBucket {
		return true
	}

	return (orgBucket == nil || orgBucket.RequestToken()) && (saBucket == nil || saBucket.RequestToken())
}

func (mrl *MemoryRateLimiter) StartSync(ctx context.Context, interval time.Duration, findRateLimits func() ([]*model.RateLimit, error)) {
	mrl.syncCache(ctx, findRateLimits)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mrl.syncCache(ctx, findRateLimits)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (mrl *MemoryRateLimiter) orgBucket(orgID int) (*Bucket, bool) {
	mrl.rw.RLock()
	defer mrl.rw.RUnlock()

	bucket, ok := mrl.orgLimits[orgID]
	return bucket, ok
}

func (mrl *MemoryRateLimiter) saBucket(saID int) (*Bucket, bool) {
	mrl.rw.RLock()
	defer mrl.rw.RUnlock()

	bucket, ok := mrl.saLimits[saID]
	return bucket, ok
}

func (mrl *MemoryRateLimiter) syncCache(ctx context.Context, findRateLimits func() ([]*model.RateLimit, error)) {
	slog.Info("started rate limit cache sync...")

	limits, err := findRateLimits()

	if err != nil {
		slog.Error("error finding rate limits", "err", err)
		return
	}

	orgLimitsMap := make(map[int]struct{})
	saLimitsMap := make(map[int]struct{})

	mrl.rw.Lock()
	defer mrl.rw.Unlock()

	for _, limit := range limits {
		if limit.ServiceAccountID == nil {
			orgLimitsMap[limit.OrgID] = struct{}{}

			if _, ok := mrl.orgLimits[limit.OrgID]; !ok {
				mrl.orgLimits[limit.OrgID] = NewBucket(limit.LimitPerMinute)
				mrl.orgLimits[limit.OrgID].Start(ctx)
			} else {
				mrl.orgLimits[limit.OrgID].UpdateCapacity(limit.LimitPerMinute)
			}
		} else {
			saLimitsMap[*limit.ServiceAccountID] = struct{}{}

			if _, ok := mrl.saLimits[*limit.ServiceAccountID]; !ok {
				mrl.saLimits[*limit.ServiceAccountID] = NewBucket(limit.LimitPerMinute)
				mrl.saLimits[*limit.ServiceAccountID].Start(ctx)
			} else {
				mrl.saLimits[*limit.ServiceAccountID].UpdateCapacity(limit.LimitPerMinute)
			}
		}
	}

	for k, v := range mrl.orgLimits {
		if _, ok := orgLimitsMap[k]; !ok {
			v.Stop()
			delete(mrl.orgLimits, k)
		}
	}

	for k, v := range mrl.saLimits {
		if _, ok := saLimitsMap[k]; !ok {
			v.Stop()
			delete(mrl.saLimits, k)
		}
	}

	slog.Info("finished rate limit cache sync...")
}
