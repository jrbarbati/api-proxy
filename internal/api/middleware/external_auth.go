package middleware

import "net/http"

func ExternalAuth(jwtSigningSecret string) func(http.Handler) http.Handler {
	return HandleAuth(jwtSigningSecret, "external")
}
