package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	bearer = "Bearer "
)

func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return "", errors.New("no auth header")
	}

	if !strings.HasPrefix(authHeader, bearer) {
		return "", errors.New("invalid auth header")
	}

	return strings.TrimPrefix(authHeader, bearer), nil
}

func VerifyJWT(token, jwtSigningSecret string) (jwt.MapClaims, error) {
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
