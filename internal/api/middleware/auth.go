package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	bearer              = "Bearer "
	orgIdKey contextKey = "org_id"
)

func handleAuth(jwtSigningSecret, desiredTokenType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r)

			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			claims, err := verifyJWT(token, jwtSigningSecret)

			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			tokenType, ok := claims["type"]

			if !ok || tokenType != desiredTokenType {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := r.Context()

			if orgId, ok := claims["org_id"]; ok {
				ctx = context.WithValue(ctx, orgIdKey, orgId)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return "", errors.New("no auth header")
	}

	if !strings.HasPrefix(authHeader, bearer) {
		return "", errors.New("invalid auth header")
	}

	return strings.TrimPrefix(authHeader, bearer), nil
}

func verifyJWT(token, jwtSigningSecret string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(jwtSigningSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := parsed.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
