package middleware

import (
	"api-proxy/internal/model"
	"log/slog"
	"net/http"
	"time"
)

type Logger interface {
	Log(route *model.Route, method, url string, statusCode int, latency time.Duration)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func LogRequest(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			r = NewRouteHolder(r)

			start := time.Now()
			next.ServeHTTP(rr, r)
			end := time.Now()

			slog.Info(
				"Request completed.",
				"method",
				r.Method,
				"path",
				r.URL.Path,
				"status",
				rr.statusCode,
				"response_time",
				end.Sub(start).Milliseconds(),
			)

			route := MatchedRoute(r)

			logger.Log(route, r.Method, r.URL.Path, rr.statusCode, end.Sub(start))
		})
	}
}
