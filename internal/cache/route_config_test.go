package cache

import (
	"api-proxy/internal/model"
	"cmp"
	"context"
	"errors"
	"net/http"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestRouteCache_FindActiveByFilter(t *testing.T) {
	now := time.Now()

	cache := RouteCache{
		rw: sync.RWMutex{},
		cache: map[string]*model.Route{
			"/api/v1/orgs:GET": {
				ID: 1, Pattern: "/api/v1/orgs", Method: http.MethodGet, UpdatedAt: new(now),
			},
			"/api/v1/orgs:POST": {
				ID: 2, Pattern: "/api/v1/orgs", Method: http.MethodPost, UpdatedAt: new(now.AddDate(0, 0, -1)),
			},
			"/api/v1/orgs:PUT": {
				ID: 3, Pattern: "/api/v1/users", Method: http.MethodPut, UpdatedAt: new(now.AddDate(0, 0, 1)),
			},
			"/api/v1/orgs:DELETE": {
				ID: 4, Pattern: "/api/v1/users", Method: http.MethodDelete, UpdatedAt: new(now.AddDate(0, 0, -1)),
			},
		},
	}
	scenarios := []struct {
		name           string
		filter         *model.RouteFilter
		cache          *RouteCache
		expectedRoutes []*model.Route
	}{
		{
			name:   "all",
			filter: &model.RouteFilter{},
			cache:  new(cache),
			expectedRoutes: []*model.Route{
				{
					ID: 1, Pattern: "/api/v1/orgs", Method: http.MethodGet, UpdatedAt: new(now),
				},
				{
					ID: 2, Pattern: "/api/v1/orgs", Method: http.MethodPost, UpdatedAt: new(now.AddDate(0, 0, -1)),
				},
				{
					ID: 3, Pattern: "/api/v1/users", Method: http.MethodPut, UpdatedAt: new(now.AddDate(0, 0, 1)),
				},
				{
					ID: 4, Pattern: "/api/v1/users", Method: http.MethodDelete, UpdatedAt: new(now.AddDate(0, 0, -1)),
				},
			},
		},
		{
			name: "pattern",
			filter: &model.RouteFilter{
				Pattern: "/api/v1/orgs",
			},
			cache: new(cache),
			expectedRoutes: []*model.Route{
				{
					ID: 1, Pattern: "/api/v1/orgs", Method: http.MethodGet, UpdatedAt: new(now),
				},
				{
					ID: 2, Pattern: "/api/v1/orgs", Method: http.MethodPost, UpdatedAt: new(now.AddDate(0, 0, -1)),
				},
			},
		},
		{
			name: "method",
			filter: &model.RouteFilter{
				Method: http.MethodDelete,
			},
			cache: new(cache),
			expectedRoutes: []*model.Route{
				{
					ID: 4, Pattern: "/api/v1/users", Method: http.MethodDelete, UpdatedAt: new(now.AddDate(0, 0, -1)),
				},
			},
		},
		{
			name: "update before",
			filter: &model.RouteFilter{
				UpdatedBefore: new(now),
			},
			cache: new(cache),
			expectedRoutes: []*model.Route{
				{
					ID: 2, Pattern: "/api/v1/orgs", Method: http.MethodPost, UpdatedAt: new(now.AddDate(0, 0, -1)),
				},
				{
					ID: 4, Pattern: "/api/v1/users", Method: http.MethodDelete, UpdatedAt: new(now.AddDate(0, 0, -1)),
				},
			},
		},
		{
			name: "update after",
			filter: &model.RouteFilter{
				UpdatedAfter: new(now),
			},
			cache: new(cache),
			expectedRoutes: []*model.Route{
				{
					ID: 3, Pattern: "/api/v1/users", Method: http.MethodPut, UpdatedAt: new(now.AddDate(0, 0, 1)),
				},
			},
		},
		{
			name: "pattern and method",
			filter: &model.RouteFilter{
				Pattern: "/api/v1/orgs",
				Method:  http.MethodPost,
			},
			cache: new(cache),
			expectedRoutes: []*model.Route{
				{
					ID: 2, Pattern: "/api/v1/orgs", Method: http.MethodPost, UpdatedAt: new(now.AddDate(0, 0, -1)),
				},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			routes, err := cache.FindActiveByFilter(scenario.filter)

			sort.Slice(routes, func(i, j int) bool {
				return cmp.Compare(routes[i].ID, routes[j].ID) < 0
			})

			sort.Slice(scenario.expectedRoutes, func(i, j int) bool {
				return cmp.Compare(routes[i].ID, routes[j].ID) < 0
			})

			if err != nil {
				t.Fatalf("FindActiveByFilter() error = %v", err)
			}

			if !reflect.DeepEqual(routes, scenario.expectedRoutes) {
				t.Fatalf("FindActiveByFilter() got = %v, want %v", routes, scenario.expectedRoutes)
			}
		})
	}
}

func TestRouteCache_Get(t *testing.T) {
	scenarios := []struct {
		name          string
		search        string
		cache         *RouteCache
		expectFound   bool
		expectedRoute *model.Route
	}{
		{
			name:   "cache hit",
			search: "/api/v1/orgs:GET",
			cache: &RouteCache{
				rw: sync.RWMutex{},
				cache: map[string]*model.Route{
					"/api/v1/orgs:GET": {
						ID: 1,
					},
				},
			},
			expectFound: true,
			expectedRoute: &model.Route{
				ID: 1,
			},
		},
		{
			name:   "cache miss",
			search: "/api/v1/orgs:POST",
			cache: &RouteCache{
				rw: sync.RWMutex{},
				cache: map[string]*model.Route{
					"/api/v1/orgs:GET": {
						ID: 1,
					},
				},
			},
			expectFound:   false,
			expectedRoute: nil,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			route, found := scenario.cache.Get(scenario.search)

			if scenario.expectFound != found {
				t.Fatalf("expected %v, got %v", scenario.expectFound, found)
			}

			if !reflect.DeepEqual(scenario.expectedRoute, route) {
				t.Fatalf("expected %v, got %v", scenario.expectedRoute, route)
			}
		})
	}
}

func TestRouteCache_Set(t *testing.T) {
	scenarios := []struct {
		name          string
		addKey        string
		addRoute      *model.Route
		startingCache *RouteCache
		endingCache   *RouteCache
	}{
		{
			name:   "set",
			addKey: "/api/v1/orgs:GET",
			addRoute: &model.Route{
				ID: 1,
			},
			startingCache: &RouteCache{
				rw:    sync.RWMutex{},
				cache: map[string]*model.Route{},
			},
			endingCache: &RouteCache{
				rw: sync.RWMutex{},
				cache: map[string]*model.Route{
					"/api/v1/orgs:GET": {
						ID: 1,
					},
				},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			scenario.startingCache.Set(scenario.addKey, scenario.addRoute)

			if !reflect.DeepEqual(scenario.endingCache.cache, scenario.startingCache.cache) {
				t.Fatalf("expected %v, got %v", scenario.endingCache, scenario.startingCache)
			}
		})
	}
}

func TestRouteCache_Delete(t *testing.T) {
	scenarios := []struct {
		name          string
		removeKey     string
		startingCache *RouteCache
		endingCache   *RouteCache
	}{
		{
			name:      "delete",
			removeKey: "/api/v1/orgs:GET",
			startingCache: &RouteCache{
				rw: sync.RWMutex{},
				cache: map[string]*model.Route{
					"/api/v1/orgs:GET":  {},
					"/api/v1/orgs:POST": {},
				},
			},
			endingCache: &RouteCache{
				rw: sync.RWMutex{},
				cache: map[string]*model.Route{
					"/api/v1/orgs:POST": {},
				},
			},
		},
		{
			name:      "delete missing key",
			removeKey: "/api/v1/orgs:PUT",
			startingCache: &RouteCache{
				rw: sync.RWMutex{},
				cache: map[string]*model.Route{
					"/api/v1/orgs:GET":  {},
					"/api/v1/orgs:POST": {},
				},
			},
			endingCache: &RouteCache{
				rw: sync.RWMutex{},
				cache: map[string]*model.Route{
					"/api/v1/orgs:GET":  {},
					"/api/v1/orgs:POST": {},
				},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			scenario.startingCache.Delete(scenario.removeKey)

			if !reflect.DeepEqual(scenario.endingCache.cache, scenario.startingCache.cache) {
				t.Fatalf("expected %v, got %v", scenario.endingCache, scenario.startingCache)
			}
		})
	}
}

func TestRouteCache_Clear(t *testing.T) {
	scenarios := []struct {
		name          string
		startingCache *RouteCache
	}{
		{
			name: "clear",
			startingCache: &RouteCache{
				rw: sync.RWMutex{},
				cache: map[string]*model.Route{
					"/api/v1/orgs:GET":  {},
					"/api/v1/orgs:POST": {},
					"/api/v1/orgs:PUT":  {},
				},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			scenario.startingCache.Clear()

			if len(scenario.startingCache.cache) > 0 {
				t.Fatalf("cache is not empty after Clear() call")
			}
		})
	}
}

func TestRouteCache_StartSync(t *testing.T) {
	scenarios := []struct {
		name                string
		cancelled           bool
		tickerDuration      time.Duration
		findRoutesFn        func() ([]*model.Route, error)
		expectedFnCallCount int
	}{
		{
			name:           "start",
			cancelled:      false,
			tickerDuration: 7 * time.Millisecond,
			findRoutesFn: func() ([]*model.Route, error) {
				return nil, nil
			},
			expectedFnCallCount: 2,
		},
		{
			name:           "error while finding routes",
			cancelled:      false,
			tickerDuration: 7 * time.Millisecond,
			findRoutesFn: func() ([]*model.Route, error) {
				return nil, errors.New("test error")
			},
			expectedFnCallCount: 2,
		},
		{
			name:           "cancel",
			cancelled:      true,
			tickerDuration: 7 * time.Millisecond,
			findRoutesFn: func() ([]*model.Route, error) {
				return nil, nil
			},
			expectedFnCallCount: 1,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			count := 0 // local to this iteration

			cache := NewRouteCache()
			ctx, cancel := context.WithCancel(context.Background())

			findRoutesFn := func() ([]*model.Route, error) {
				count++
				return scenario.findRoutesFn()
			}

			if !scenario.cancelled {
				cache.StartSync(ctx, scenario.tickerDuration, findRoutesFn)
				time.Sleep(10 * time.Millisecond)
				cancel()
			} else {
				cache.StartSync(ctx, scenario.tickerDuration, findRoutesFn)
				cancel()
				time.Sleep(10 * time.Millisecond)
			}

			if scenario.expectedFnCallCount != count {
				t.Fatalf("expected call count %v, got %v", scenario.expectedFnCallCount, count)
			}
		})
	}
}
