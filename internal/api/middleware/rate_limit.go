package middleware

import (
	"api-proxy/internal/model"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RateLimiter interface {
	AllowRequest(orgID, saID int) bool
	StartSync(ctx context.Context, interval time.Duration, findRateLimits func() ([]*model.RateLimit, error))
}

func RateLimit(rateLimiter RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(claimsKey) == nil {
				slog.Error("made it to rate limiter with missing auth/claims")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			var orgID int
			var serviceAccountID int
			var tokenType string
			var tokenSubType string

			claims, ok := r.Context().Value(claimsKey).(jwt.MapClaims)

			if !ok {
				slog.Error("unexpected type from \"claims\" in the context", "claims", r.Context().Value(claimsKey))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			orgIDClaim, ok := claims["org_id"].(float64)
			orgID = int(orgIDClaim)

			if !ok {
				slog.Error("org_id missing from claims or invalid", "org_id", claims["org_id"])
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			serviceAccountIDClaim, ok := claims["sub"].(float64)
			serviceAccountID = int(serviceAccountIDClaim)

			if !ok {
				slog.Error("sub missing from claims or invalid", "sub", claims["sub"])
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			tokenType, ok = claims["type"].(string)

			if !ok {
				slog.Error("type missing from claims or invalid", "type", claims["type"])
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			tokenSubType, ok = claims["sub_type"].(string)

			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			if tokenType != "external" || tokenSubType != "service-account" {
				next.ServeHTTP(w, r)
				return
			}

			if !rateLimiter.AllowRequest(orgID, serviceAccountID) {
				slog.Info("rate limiting request", "org_id", orgID, "service_account_id", serviceAccountID)
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
