package middleware

import "net/http"

func ExternalAuth(jwtSigningSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement
			next.ServeHTTP(w, r)
		})
	}
}
