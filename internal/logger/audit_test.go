package logger

import (
	"api-proxy/internal/model"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeAuditLogDataStore struct {
	auditLogs []*model.AuditLog
	err       error
}

func (f *fakeAuditLogDataStore) Insert(auditLog *model.AuditLog) (*model.AuditLog, error) {
	if f.err == nil {
		f.auditLogs = append(f.auditLogs, auditLog)
		return f.auditLogs[0], nil
	}

	return nil, f.err
}

func TestAuditLogger_Log(t *testing.T) {
	scenarios := []struct {
		name            string
		readFromChan    bool
		numLogs         int
		queueSize       int
		expectedChanLen int
		dataStore       AuditLogDataStorer
	}{
		{name: "Persisted", readFromChan: true, numLogs: 2, queueSize: 10, expectedChanLen: 2, dataStore: &fakeAuditLogDataStore{}},
		{name: "Dropped", readFromChan: false, numLogs: 1, queueSize: 0, expectedChanLen: 0, dataStore: &fakeAuditLogDataStore{}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			auditLogLogger := NewAuditLogger(scenario.dataStore, scenario.queueSize)

			for i := 0; i < scenario.numLogs; i++ {
				auditLogLogger.Log(i, model.ROUTE, 12, model.CREATE)
			}

			if scenario.expectedChanLen != len(auditLogLogger.ch) {
				t.Errorf("expected channel length to be %d, got %d", scenario.expectedChanLen, len(auditLogLogger.ch))
			}

			if scenario.readFromChan {
				for i := 0; i < scenario.numLogs; i++ {
					entry := <-auditLogLogger.ch

					if entry.EntityID != i {
						t.Errorf("expected entity id %d, got %d", i, entry.EntityID)
					}

					if entry.EntityType != "route" {
						t.Errorf("expected entity type Route, got %s", entry.EntityType)
					}
					if entry.PerformedByID != 12 {
						t.Errorf("expected performed by id 12, got %d", entry.PerformedByID)
					}
					if entry.Action != "create" {
						t.Errorf("expected action create, got %s", entry.Action)
					}
				}

				if len(auditLogLogger.ch) != 0 {
					t.Errorf("expected empty channel, got length of %d", len(auditLogLogger.ch))
				}
			}
		})
	}
}

func TestAuditLogger_Start(t *testing.T) {
	scenarios := []struct {
		name      string
		entry     Auditable
		dataStore AuditLogDataStorer
		cancelled bool
		assert    func(t *testing.T, ds *fakeAuditLogDataStore)
	}{
		{
			name:      "Persisted",
			entry:     NewAuditable(1, "Route", 12, "CREATE"),
			dataStore: &fakeAuditLogDataStore{},
			cancelled: false,
			assert: func(t *testing.T, ds *fakeAuditLogDataStore) {
				if len(ds.auditLogs) != 1 {
					t.Errorf("expected 1 persisted auditLog, got %d", len(ds.auditLogs))
				}
			},
		},
		{
			name:      "Errored",
			entry:     NewAuditable(1, "Route", 12, "CREATE"),
			dataStore: &fakeAuditLogDataStore{err: errors.New("test insert err")},
			cancelled: false,
			assert: func(t *testing.T, ds *fakeAuditLogDataStore) {
				if len(ds.auditLogs) > 0 {
					t.Errorf("expected 0 persisted auditLogs, got %d", len(ds.auditLogs))
				}
			},
		},
		{
			name:      "Cancelled",
			entry:     NewAuditable(1, "Route", 12, "CREATE"),
			dataStore: &fakeAuditLogDataStore{},
			cancelled: true,
			assert: func(t *testing.T, ds *fakeAuditLogDataStore) {
				if len(ds.auditLogs) > 0 {
					t.Errorf("expected 0 persisted auditLogs (context cancelled), got %d", len(ds.auditLogs))
				}
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			auditLogLogger := NewAuditLogger(scenario.dataStore, 1)
			ctx, cancel := context.WithCancel(context.Background())

			auditLogLogger.Start(ctx)

			if !scenario.cancelled {
				auditLogLogger.ch <- scenario.entry // blocking if nothing is reading from the channel
				time.Sleep(time.Duration(10) * time.Millisecond)
				cancel()
			} else {
				cancel()
				time.Sleep(10 * time.Millisecond)
				auditLogLogger.Log(12, model.ROUTE, 12, model.CREATE)
			}

			scenario.assert(t, scenario.dataStore.(*fakeAuditLogDataStore))
		})
	}
}
