package logger

import (
	"context"
	"log/slog"
	"time"
)

type RequestLog struct {
	Method     string
	URL        string
	StatusCode int
	Latency    time.Duration
}

type RequestLogger struct {
	ch chan RequestLog
}

func NewRequestLogger(queueSize int) RequestLogger {
	return RequestLogger{
		ch: make(chan RequestLog, queueSize),
	}
}

func NewRequestLog(method, url string, statusCode int, latency time.Duration) RequestLog {
	return RequestLog{
		Method:     method,
		URL:        url,
		StatusCode: statusCode,
		Latency:    latency,
	}
}

// Log insert log requests into the channel for asynchronous logging (if RequestLogger was configured with queueSize > 0, default is 500), synchronous logging if queueSize == 0
func (rl RequestLogger) Log(method, url string, statusCode int, latency time.Duration) {
	select {
	case rl.ch <- NewRequestLog(method, url, statusCode, latency):
	default:
		slog.Warn("request log channel full, dropping entry")
	}
}

// Start starts up the goroutine to handle log requests from the channel
func (rl RequestLogger) Start(ctx context.Context) {
	// TODO: Here instead of logging, we can do something more complicated like saving to a DB or other persistence mechanism
	slog.Info("Starting request logger...", "queue_size", cap(rl.ch))

	go func() {
		for {
			select {
			case entry := <-rl.ch:
				slog.Info(
					"Request completed.",
					"method",
					entry.Method,
					"path",
					entry.URL,
					"status",
					entry.StatusCode,
					"response_time",
					entry.Latency,
				)
			case <-ctx.Done():
				return
			}
		}
	}()
}
