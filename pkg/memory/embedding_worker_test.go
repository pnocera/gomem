package memory

import (
	"context"
	"testing"
	"time"
)

// TestNewEmbeddingWorker ensures worker can be created.
func TestNewEmbeddingWorker(t *testing.T) {
	cfg := &Config{TopicMemoryEmbed: "test.topic.embed"} // Minimal config
	mockNATS := &mockNATSClient{}
	mockOpenAI := &mockOpenAIClient{} // Re-use mock from processing_worker_test
	worker := NewEmbeddingWorker(mockNATS, cfg, mockOpenAI)
	if worker == nil {
		t.Errorf("NewEmbeddingWorker returned nil")
	}
	if worker.nc != mockNATS {
		t.Error("EmbeddingWorker: NATS client not set correctly")
	}
	if worker.cfg != cfg {
		t.Error("EmbeddingWorker: Config not set correctly")
	}
	if worker.openai != mockOpenAI {
		t.Error("EmbeddingWorker: OpenAI client not set correctly")
	}
}

// TestEmbeddingWorker_StartStop ensures Start can be called and respects context cancellation.
func TestEmbeddingWorker_StartStop(t *testing.T) {
	cfg := &Config{TopicMemoryEmbed: "test.embedding.startstop"}
	mockNATS := &mockNATSClient{}
	mockOpenAI := &mockOpenAIClient{}
	worker := NewEmbeddingWorker(mockNATS, cfg, mockOpenAI)

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
