package middleware

import (
	"api-proxy/internal/model"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type RouteStorer interface {
	All() []*model.Route
}

const matchedRouteKey contextKey = "matched_route"

var ErrRouteNotFound = errors.New("route not found")

func ResolveRoute(routeStore RouteStorer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			routes := routeStore.All()

			route, err := findRoute(routes, r)

			if err != nil {
				if errors.Is(err, ErrRouteNotFound) {
					slog.Warn("route not found", "route", route)
				}
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			ctx := context.WithValue(r.Context(), matchedRouteKey, route)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func MatchedRoute(r *http.Request) *model.Route {
	return r.Context().Value(matchedRouteKey).(*model.Route)
}

func findRoute(routes []*model.Route, r *http.Request) (*model.Route, error) {
	requestSegments := splitUri(r.URL.Path)

	for _, route := range routes {
		if string(route.Method) != r.Method {
			continue
		}

		routeSegments := splitUri(route.Pattern)

		if ok := compareSegments(requestSegments, routeSegments); !ok {
			continue
		}

		return route, nil
	}

	return nil, ErrRouteNotFound
}

func splitUri(uri string) []string {
	return strings.Split(uri, "/")
}

func compareSegments(request, possibleRoute []string) bool {
	if len(request) != len(possibleRoute) {
		return false
	}

	for i := 0; i < len(request); i++ {
		if !verifySegment(request[i], possibleRoute[i]) {
			return false
		}
	}

	return true
}

func verifySegment(requestSegment, possibleRouteSegment string) bool {
	if strings.HasPrefix(possibleRouteSegment, "{") && strings.HasSuffix(possibleRouteSegment, "}") {
		return true
	}

	return requestSegment == possibleRouteSegment
}
