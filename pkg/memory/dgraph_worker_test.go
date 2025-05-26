package memory

import (
	"context"
	"testing"
	"time"

	"github.com/pnocera/gomem/pkg/graphs" // Assuming module path for graphs.GraphStoreConfig
)

// --- Mock DgraphClient for DgraphWorker tests ---
type mockDgraphClient struct {
	MutateFunc func(ctx context.Context, data interface{}) error
	QueryFunc  func(ctx context.Context, query string, vars map[string]string) ([]byte, error)
}

func (m *mockDgraphClient) Mutate(ctx context.Context, data interface{}) error {
	if m.MutateFunc != nil {
		return m.MutateFunc(ctx, data)
	}
	return nil
}

func (m *mockDgraphClient) Query(ctx context.Context, query string, vars map[string]string) ([]byte, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, query, vars)
	}
	return nil, nil
}

// TestNewDgraphWorker ensures worker can be created.
func TestNewDgraphWorker(t *testing.T) {
	cfg := &Config{EnableGraphStore: true, TopicMemoryGraphStoreAdd: "test.topic.dgraph"} // Minimal config
	mockNATS := &mockNATSClient{}
	mockOpenAI := &mockOpenAIClient{} // Re-use from other worker tests
	mockDG := &mockDgraphClient{}
	mockGraphCfg := &graphs.GraphStoreConfig{}

	worker := NewDgraphWorker(mockNATS, cfg, mockOpenAI, mockDG, mockGraphCfg)
	if worker == nil {
		t.Errorf("NewDgraphWorker returned nil")
	}
	if worker.nc != mockNATS {
		t.Error("DgraphWorker: NATS client not set correctly")
	}
	if worker.cfg != cfg {
		t.Error("DgraphWorker: Config not set correctly")
	}
	if worker.openai != mockOpenAI {
		t.Error("DgraphWorker: OpenAI client not set correctly")
	}
	if worker.dg != mockDG {
		t.Error("DgraphWorker: Dgraph client not set correctly")
	}
	if worker.graphCfg != mockGraphCfg {
		t.Error("DgraphWorker: GraphStoreConfig not set correctly")
	}
}

// TestDgraphWorker_StartStop ensures Start can be called and respects context cancellation.
func TestDgraphWorker_StartStop(t *testing.T) {
	cfgEnabled := &Config{EnableGraphStore: true, TopicMemoryGraphStoreAdd: "test.dgraph.startstop.enabled"}
	cfgDisabled := &Config{EnableGraphStore: false, TopicMemoryGraphStoreAdd: "test.dgraph.startstop.disabled"}
	mockNATS := &mockNATSClient{}
	mockOpenAI := &mockOpenAIClient{}
	mockDG := &mockDgraphClient{}
	mockGraphCfg := &graphs.GraphStoreConfig{}

	t.Run("Enabled", func(t *testing.T) {
		worker := NewDgraphWorker(mockNATS, cfgEnabled, mockOpenAI, mockDG, mockGraphCfg)
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
	})

	t.Run("Disabled", func(t *testing.T) {
		worker := NewDgraphWorker(mockNATS, cfgDisabled, mockOpenAI, mockDG, mockGraphCfg)
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond) // Increased timeout
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- worker.Start(ctx)
		}()
		// Worker should also stop if disabled
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("Worker Start (disabled) returned unexpected error: %v, expected nil", err)
			}
		case <-time.After(400 * time.Millisecond): // Increased test safety timeout
			t.Errorf("Worker Start (disabled) did not return after context cancellation")
		}
	})
}
