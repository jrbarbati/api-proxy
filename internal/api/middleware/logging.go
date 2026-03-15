package middleware

import (
	"net/http"
	"time"
)

type Logger interface {
	Log(method, url string, statusCode int, latency time.Duration)
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

			start := time.Now()
			next.ServeHTTP(rr, r)
			end := time.Now()

			logger.Log(r.Method, r.URL.Path, rr.statusCode, end.Sub(start))
		})
	}
}
