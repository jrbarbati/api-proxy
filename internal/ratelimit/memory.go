package ratelimit

import (
	"api-proxy/internal/model"
	"context"
	"log/slog"
	"sync"
	"time"
)

type tokenBucket interface {
	Start(context.Context)
	Stop()
	UpdateCapacity(newCapacity int)
	requestToken() bool
}

type MemoryRateLimiter struct {
	rw             sync.RWMutex
	orgLimits      map[int]tokenBucket
	saLimits       map[int]tokenBucket
	newTokenBucket func(int) tokenBucket
}

func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		rw:        sync.RWMutex{},
		orgLimits: make(map[int]tokenBucket),
		saLimits:  make(map[int]tokenBucket),
		newTokenBucket: func(limit int) tokenBucket {
			return newBucket(limit)
		},
	}
}

func (mrl *MemoryRateLimiter) AllowRequest(orgID, saID int) bool {
	orgBucket, orgHasBucket := mrl.orgBucket(orgID)
	saBucket, saHasBucket := mrl.saBucket(saID)

	if !orgHasBucket && !saHasBucket {
		return true
	}

	return (orgBucket == nil || orgBucket.requestToken()) && (saBucket == nil || saBucket.requestToken())
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

func (mrl *MemoryRateLimiter) orgBucket(orgID int) (tokenBucket, bool) {
	mrl.rw.RLock()
	defer mrl.rw.RUnlock()

	b, ok := mrl.orgLimits[orgID]
	return b, ok
}

func (mrl *MemoryRateLimiter) saBucket(saID int) (tokenBucket, bool) {
	mrl.rw.RLock()
	defer mrl.rw.RUnlock()

	b, ok := mrl.saLimits[saID]
	return b, ok
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
				mrl.orgLimits[limit.OrgID] = mrl.newTokenBucket(limit.LimitPerMinute)
				mrl.orgLimits[limit.OrgID].Start(ctx)
			} else {
				mrl.orgLimits[limit.OrgID].UpdateCapacity(limit.LimitPerMinute)
			}
		} else {
			saLimitsMap[*limit.ServiceAccountID] = struct{}{}

			if _, ok := mrl.saLimits[*limit.ServiceAccountID]; !ok {
				mrl.saLimits[*limit.ServiceAccountID] = mrl.newTokenBucket(limit.LimitPerMinute)
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
