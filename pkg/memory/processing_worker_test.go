package memory

import (
	"context"
	"testing"
	"time"
)

// --- Mock OpenAIClient for worker tests ---
type mockOpenAIClient struct {
	ExtractFactsReturn     string
	ExtractFactsError      error
	GetEmbeddingReturn     []float32
	GetEmbeddingError      error
	ExtractGraphDataReturn struct {
		Entities  []Entity
		Relations []Relation
	}
	ExtractGraphDataError error
}

func (m *mockOpenAIClient) ExtractFacts(ctx context.Context, text []string, prompt string) (string, error) {
	return m.ExtractFactsReturn, m.ExtractFactsError
}
func (m *mockOpenAIClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return m.GetEmbeddingReturn, m.GetEmbeddingError
}
func (m *mockOpenAIClient) ExtractGraphData(ctx context.Context, text string, prompt string) ([]Entity, []Relation, error) {
	return m.ExtractGraphDataReturn.Entities, m.ExtractGraphDataReturn.Relations, m.ExtractGraphDataError
}

// TestNewProcessingWorker ensures worker can be created.
func TestNewProcessingWorker(t *testing.T) {
	cfg := &Config{TopicMemoryProcess: "test.topic"} // Minimal config for constructor
	mockNATS := &mockNATSClient{}
	mockOpenAI := &mockOpenAIClient{}
	worker := NewProcessingWorker(mockNATS, cfg, mockOpenAI)
	if worker == nil {
		t.Errorf("NewProcessingWorker returned nil")
	}
	if worker.nc != mockNATS {
		t.Error("ProcessingWorker: NATS client not set correctly")
	}
	if worker.cfg != cfg {
		t.Error("ProcessingWorker: Config not set correctly")
	}
	if worker.openai != mockOpenAI {
		t.Error("ProcessingWorker: OpenAI client not set correctly")
	}
}

// TestProcessingWorker_StartStop ensures Start can be called and respects context cancellation.
func TestProcessingWorker_StartStop(t *testing.T) {
	cfg := &Config{TopicMemoryProcess: "test.processing.startstop"} // Ensure unique topic for safety
	mockNATS := &mockNATSClient{}
	mockOpenAI := &mockOpenAIClient{}
	worker := NewProcessingWorker(mockNATS, cfg, mockOpenAI)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond) 
	
	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Start(ctx)
	}()

	// Wait for the context to be done, then check the error channel.
	<-ctx.Done() // This blocks until the timeout of 200ms passes or cancel() is called.
	cancel()     // Ensure cancellation is signaled promptly.

	select {
	case err := <-errCh: // Check what worker.Start() returned.
		if err != nil {
			// The shell Start() returns nil even on context cancellation.
			t.Errorf("Worker Start returned unexpected error: %v, expected nil", err)
		}
		// If nil, it means Start() exited cleanly as expected.
	case <-time.After(100 * time.Millisecond): // Additional short timeout to get error from channel.
		t.Errorf("Worker Start did not send return value to channel after context cancellation")
	}
}
