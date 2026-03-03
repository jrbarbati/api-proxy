package cache

import (
	"api-proxy/internal/model"
	"context"
	"log/slog"
	"sync"
	"time"
)

type RouteCache struct {
	rw       sync.RWMutex
	cache    map[string]*model.Route
	lastSync *time.Time
}

func NewRouteCache() *RouteCache {
	return &RouteCache{
		rw:       sync.RWMutex{},
		cache:    newCache(0),
		lastSync: nil,
	}
}

func (r *RouteCache) All() []*model.Route {
	r.rw.RLock()
	defer r.rw.RUnlock()

	routes := make([]*model.Route, 0, len(r.cache))

	for _, route := range r.cache {
		routes = append(routes, route)
	}

	return routes
}

func (r *RouteCache) Get(key string) (*model.Route, bool) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	item, ok := r.cache[key]

	if !ok {
		return nil, false
	}

	return item, true
}

func (r *RouteCache) Set(key string, route *model.Route) {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.cache[key] = route
}

func (r *RouteCache) Delete(key string) {
	r.rw.Lock()
	defer r.rw.Unlock()

	delete(r.cache, key)
}

func (r *RouteCache) Clear() {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.cache = newCache(len(r.cache))
}

func newCache(size int) map[string]*model.Route {
	return make(map[string]*model.Route, size)
}

func (r *RouteCache) StartSync(ctx context.Context, interval time.Duration, findRoutes func() ([]*model.Route, error)) {
	r.syncCache(findRoutes)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.syncCache(findRoutes)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (r *RouteCache) syncCache(findRoutes func() ([]*model.Route, error)) {
	slog.Info("started sync cache...")

	routes, err := findRoutes()

	if err != nil {
		slog.Error("error syncing routes from db to cache: ", err)
		return
	}

	nc := newCache(len(routes))

	for _, route := range routes {
		nc[route.Pattern+":"+string(route.Method)] = route
	}

	r.rw.Lock()
	defer r.rw.Unlock()
	r.cache = nc

	slog.Info("finished sync cache...", "routesAdded", len(routes))
}
