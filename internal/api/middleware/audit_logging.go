package middleware

import (
	"api-proxy/internal/model"
	"log/slog"
	"net/http"
	"strconv"
)

type AuditLogger interface {
	Log(entityID int, entityType model.EntityType, performedById int, action model.Action)
}

func LogAuditable(auditLogger AuditLogger, entityType model.EntityType, action model.Action) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Get the Entity ID from somewhere, maybe mirror the response recorder in the other logging middleware
			next.ServeHTTP(w, r)

			claims := Claims(r)

			userIdStr, err := claims.GetSubject()

			if err != nil {
				slog.Error("error getting subject from token claims", "err", err)
				return
			}

			userId, err := strconv.Atoi(userIdStr)

			if err != nil {
				slog.Error("error converting user id from token claims", "err", err)
				return
			}

			auditLogger.Log(10, entityType, userId, action)
		})
	}
}
