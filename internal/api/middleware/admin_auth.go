package middleware

import "net/http"

func AdminAuth(jwtSigningSecret string) func(http.Handler) http.Handler {
	return HandleAuth(jwtSigningSecret, "internal")
}
