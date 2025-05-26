package memory

import (
	"context"
	"testing"
	"time"
)

// TestNewHistoryWorker ensures worker can be created.
func TestNewHistoryWorker(t *testing.T) {
	cfg := &Config{TopicMemoryHistoryLog: "test.topic.history"} // Minimal config
	mockNATS := &mockNATSClient{}
	mockHistory := &mockHistoryStore{} // Re-use mock from service_test.go
	worker := NewHistoryWorker(mockNATS, cfg, mockHistory)
	if worker == nil {
		t.Errorf("NewHistoryWorker returned nil")
	}
	if worker.nc != mockNATS {
		t.Error("HistoryWorker: NATS client not set correctly")
	}
	if worker.cfg != cfg {
		t.Error("HistoryWorker: Config not set correctly")
	}
	if worker.historyStore != mockHistory {
		t.Error("HistoryWorker: HistoryStore not set correctly")
	}
}

// TestHistoryWorker_StartStop ensures Start can be called and respects context cancellation.
func TestHistoryWorker_StartStop(t *testing.T) {
	cfg := &Config{TopicMemoryHistoryLog: "test.history.startstop"}
	mockNATS := &mockNATSClient{}
	mockHistory := &mockHistoryStore{}
	worker := NewHistoryWorker(mockNATS, cfg, mockHistory)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond) // Increased timeout
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Start(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Worker Start returned unexpected error: %v, expected nil on context done", err)
		}
	case <-time.After(400 * time.Millisecond): // Increased test safety timeout
		t.Errorf("Worker Start did not return after context cancellation")
	}
}
