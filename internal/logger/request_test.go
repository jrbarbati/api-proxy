package logger

import (
	"api-proxy/internal/model"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeDataStore struct {
	requests []*model.Request
	err      error
}

func (f *fakeDataStore) Insert(request *model.Request) (*model.Request, error) {
	if f.err == nil {
		f.requests = append(f.requests, request)
		return f.requests[0], nil
	}

	return nil, f.err
}

func TestRequestLogger_Log(t *testing.T) {
	scenarios := []struct {
		name            string
		readFromChan    bool
		numLogs         int
		queueSize       int
		expectedChanLen int
		dataStore       RequestDataStorer
	}{
		{name: "Persisted", readFromChan: true, numLogs: 2, queueSize: 10, expectedChanLen: 2, dataStore: &fakeDataStore{}},
		{name: "Dropped", readFromChan: false, numLogs: 1, queueSize: 0, expectedChanLen: 0, dataStore: &fakeDataStore{}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			requestLogger := NewRequestLogger(scenario.dataStore, scenario.queueSize)

			for i := 0; i < scenario.numLogs; i++ {
				requestLogger.Log("GET", "/api/v1/test", 201, time.Duration(200)*time.Millisecond)
			}

			if scenario.expectedChanLen != len(requestLogger.ch) {
				t.Errorf("expected channel length to be %d, got %d", scenario.expectedChanLen, len(requestLogger.ch))
			}

			if scenario.readFromChan {
				for i := 0; i < scenario.numLogs; i++ {
					entry := <-requestLogger.ch

					if entry.Method != "GET" {
						t.Errorf("expected method GET, got %s", entry.Method)
					}
					if entry.URL != "/api/v1/test" {
						t.Errorf("expected url /api/v1/test, got %s", entry.URL)
					}
					if entry.StatusCode != 201 {
						t.Errorf("expected status code 201, got %d", entry.StatusCode)
					}
					if entry.Latency.Milliseconds() != 200 {
						t.Errorf("expected latency 200ms, got %d", entry.Latency.Milliseconds())
					}
				}

				if len(requestLogger.ch) != 0 {
					t.Errorf("expected empty channel, got length of %d", len(requestLogger.ch))
				}
			}
		})
	}
}

func TestRequestLogger_Start(t *testing.T) {
	scenarios := []struct {
		name      string
		entry     RequestLog
		dataStore RequestDataStorer
		cancelled bool
		assert    func(t *testing.T, ds *fakeDataStore)
	}{
		{
			name:      "Persisted",
			entry:     NewRequestLog("GET", "/api/v1/test", 201, time.Duration(100)*time.Millisecond),
			dataStore: &fakeDataStore{},
			cancelled: false,
			assert: func(t *testing.T, ds *fakeDataStore) {
				if len(ds.requests) != 1 {
					t.Errorf("expected 1 persisted request, got %d", len(ds.requests))
				}
			},
		},
		{
			name:      "Errored",
			entry:     NewRequestLog("GET", "/api/v1/test", 201, time.Duration(100)*time.Millisecond),
			dataStore: &fakeDataStore{err: errors.New("test insert err")},
			cancelled: false,
			assert: func(t *testing.T, ds *fakeDataStore) {
				if len(ds.requests) > 0 {
					t.Errorf("expected 0 persisted requests, got %d", len(ds.requests))
				}
			},
		},
		{
			name:      "Cancelled",
			entry:     NewRequestLog("GET", "/api/v1/test", 201, time.Duration(100)*time.Millisecond),
			dataStore: &fakeDataStore{},
			cancelled: true,
			assert: func(t *testing.T, ds *fakeDataStore) {
				if len(ds.requests) > 0 {
					t.Errorf("expected 0 persisted requests (context cancelled), got %d", len(ds.requests))
				}
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			requestLogger := NewRequestLogger(scenario.dataStore, 1)
			ctx, cancel := context.WithCancel(context.Background())

			requestLogger.Start(ctx)

			if !scenario.cancelled {
				requestLogger.ch <- scenario.entry // blocking if nothing is reading from the channel
				time.Sleep(time.Duration(10) * time.Millisecond)
				cancel()
			} else {
				cancel()
				time.Sleep(10 * time.Millisecond)
				requestLogger.Log("GET", "/api/v1/test", 201, 100*time.Millisecond) // non-blocking
			}

			scenario.assert(t, scenario.dataStore.(*fakeDataStore))
		})
	}
}

// Essentially tests how long it takes for Log to put something into the channel, as the channel is async (buffered channel)
// so we do not wait for it to respond to us.
func BenchmarkRequestLogger_Log(b *testing.B) {
	b.ReportAllocs()

	requestLogger := NewRequestLogger(&fakeDataStore{}, b.N)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	requestLogger.Start(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		requestLogger.Log("GET", "/api/v1/test", 200, time.Millisecond)
	}
}
