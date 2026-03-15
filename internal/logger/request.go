package logger

import (
	"api-proxy/internal/model"
	"context"
	"log/slog"
	"time"
)

type RequestDataStorer interface {
	Insert(request *model.Request) (*model.Request, error)
}

type RequestLog struct {
	Method     string
	URL        string
	StatusCode int
	Latency    time.Duration
}

type RequestLogger struct {
	ch        chan RequestLog
	dataStore RequestDataStorer
}

func NewRequestLogger(requestDataStore RequestDataStorer, queueSize int) RequestLogger {
	return RequestLogger{
		ch:        make(chan RequestLog, queueSize),
		dataStore: requestDataStore,
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
	slog.Info("Starting request logger...", "queue_size", cap(rl.ch))

	go func() {
		for {
			select {
			case entry := <-rl.ch:
				if _, err := rl.dataStore.Insert(&model.Request{
					Method:     entry.Method,
					URL:        entry.URL,
					StatusCode: entry.StatusCode,
					Latency:    entry.Latency,
				}); err != nil {
					slog.Error("Failed to insert request", "method", entry.Method, "url", entry.URL)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
