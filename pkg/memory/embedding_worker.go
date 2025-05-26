package memory

import (
	"context"
	"encoding/json"
	"fmt"
	// Required for MemoryEvent
	// "github.com/google/uuid" // Required for MemoryEvent
)

// EmbeddingWorker handles generating embeddings for processed memories.
type EmbeddingWorker struct {
	nc     NATSClient
	cfg    *Config
	openai OpenAIClient
}

// NewEmbeddingWorker creates a new EmbeddingWorker.
func NewEmbeddingWorker(nc NATSClient, cfg *Config, openai OpenAIClient) *EmbeddingWorker {
	return &EmbeddingWorker{
		nc:     nc,
		cfg:    cfg,
		openai: openai,
	}
}

// Start begins the worker's NATS subscription.
func (w *EmbeddingWorker) Start(ctx context.Context) error {
	if w.nc == nil {
		fmt.Println("EmbeddingWorker: NATS client is nil, worker will not start.")
		<-ctx.Done()
		return nil
	}

	fmt.Printf("EmbeddingWorker started, listening on topic: %s\n", w.cfg.TopicMemoryEmbed)
	// In a real implementation, w.nc.Subscribe would be called here.
	// The handler would be w.handleEmbedMessage.
	// For shell, we simulate by just blocking.
	go func() {
		// Simulated subscription loop
	}()

	<-ctx.Done()
	fmt.Println("EmbeddingWorker shutting down.")
	return nil
}

// handleEmbedMessage simulates processing an incoming NATS message for embedding.
func (w *EmbeddingWorker) handleEmbedMessage(payload []byte) error {
	fmt.Printf("EmbeddingWorker received payload: %s\n", string(payload))

	var processedData ProcessedMemoryData
	if err := json.Unmarshal(payload, &processedData); err != nil {
		fmt.Printf("EmbeddingWorker: Error unmarshalling ProcessedMemoryData: %v\n", err)
		return fmt.Errorf("error unmarshalling ProcessedMemoryData: %w", err)
	}
	fmt.Printf("EmbeddingWorker: Unmarshalled ProcessedMemoryData for MemoryID: %s\n", processedData.MemoryID)

	var embedding []float32
	var err error
	if w.openai != nil {
		fmt.Println("EmbeddingWorker: Simulating OpenAI GetEmbedding call...")
		embedding, err = w.openai.GetEmbedding(context.Background(), processedData.ProcessedText)
		if err != nil {
			fmt.Printf("EmbeddingWorker: Error simulating OpenAI GetEmbedding: %v\n", err)
			// Decide if this is a fatal error
			return fmt.Errorf("error getting embedding: %w", err)
		}
		fmt.Printf("EmbeddingWorker: Simulated embedding generation for MemoryID: %s\n", processedData.MemoryID)
	} else {
		fmt.Println("EmbeddingWorker: OpenAI client is nil, cannot generate embedding.")
		// Use a dummy embedding if openai client is nil, or return error
		embedding = []float32{0.0, 0.0, 0.0} // Dummy embedding
	}

	embeddingData := EmbeddingData{
		BaseRequestInfo: processedData.BaseRequestInfo,
		MemoryID:        processedData.MemoryID,
		TextToEmbed:     processedData.ProcessedText, // Or specific parts if logic changes
		Embedding:       embedding,
		ProcessedText:   processedData.ProcessedText,
	}

	jsonData, err := json.Marshal(embeddingData)
	if err != nil {
		fmt.Printf("EmbeddingWorker: Error marshalling EmbeddingData: %v\n", err)
		return fmt.Errorf("error marshalling EmbeddingData: %w", err)
	}

	// Simulate publishing to TopicMemoryVectorStoreAdd
	if w.nc != nil {
		err = w.nc.Publish(context.Background(), w.cfg.TopicMemoryVectorStoreAdd, jsonData)
		if err != nil {
			fmt.Printf("EmbeddingWorker: Error publishing EmbeddingData to NATS topic %s: %v\n", w.cfg.TopicMemoryVectorStoreAdd, err)
		} else {
			fmt.Printf("EmbeddingWorker: Published EmbeddingData to %s for MemoryID: %s\n", w.cfg.TopicMemoryVectorStoreAdd, processedData.MemoryID)
		}
	} else {
		fmt.Printf("NATS_PUBLISH (EmbeddingWorker - nc is nil): Topic=%s, Payload=%s\n", w.cfg.TopicMemoryVectorStoreAdd, string(jsonData))
	}

	if w.cfg.EnableGraphStore {
		graphStoreData := GraphStoreStorageData{
			BaseRequestInfo: processedData.BaseRequestInfo,
			MemoryID:        processedData.MemoryID,
			TextForGraph:    processedData.ProcessedText, // Or specific parts
			// Entities and Relationships would be populated by a graph extraction step,
			// which happens in DgraphWorker based on the prompt for DgraphWorker.
			// So, EmbeddingWorker likely just passes the text through for graph processing.
		}
		graphJsonData, err := json.Marshal(graphStoreData)
		if err != nil {
			fmt.Printf("EmbeddingWorker: Error marshalling GraphStoreStorageData: %v\n", err)
			// Non-fatal for the main embedding flow perhaps
		} else {
			if w.nc != nil {
				err = w.nc.Publish(context.Background(), w.cfg.TopicMemoryGraphStoreAdd, graphJsonData)
				if err != nil {
					fmt.Printf("EmbeddingWorker: Error publishing GraphStoreStorageData to NATS topic %s: %v\n", w.cfg.TopicMemoryGraphStoreAdd, err)
				} else {
					fmt.Printf("EmbeddingWorker: Published GraphStoreStorageData to %s for MemoryID: %s\n", w.cfg.TopicMemoryGraphStoreAdd, processedData.MemoryID)
				}
			} else {
				fmt.Printf("NATS_PUBLISH (EmbeddingWorker - nc is nil): Topic=%s, Payload=%s\n", w.cfg.TopicMemoryGraphStoreAdd, string(graphJsonData))
			}
		}
	}
	return nil
}
