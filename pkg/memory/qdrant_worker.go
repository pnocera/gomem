package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pnocera/gomem/pkg/vectorstores"

	"github.com/google/uuid"
)

// QdrantWorker handles storing embeddings in Qdrant.
type QdrantWorker struct {
	nc  NATSClient
	cfg *Config
	vs  vectorstores.VectorStore
}

// NewQdrantWorker creates a new QdrantWorker.
func NewQdrantWorker(nc NATSClient, cfg *Config, vs vectorstores.VectorStore) *QdrantWorker {
	return &QdrantWorker{
		nc:  nc,
		cfg: cfg,
		vs:  vs,
	}
}

// Start begins the worker's NATS subscription.
func (w *QdrantWorker) Start(ctx context.Context) error {
	if w.nc == nil {
		fmt.Println("QdrantWorker: NATS client is nil, worker will not start.")
		<-ctx.Done()
		return nil
	}
	if w.vs == nil {
		fmt.Println("QdrantWorker: VectorStore client (vs) is nil, worker will not start effectively.")
		// Depending on requirements, may still start to listen but log errors in handler.
		// For shell, let's proceed but note it.
	}

	fmt.Printf("QdrantWorker started, listening on topic: %s\n", w.cfg.TopicMemoryVectorStoreAdd)
	// In a real implementation, w.nc.Subscribe would be called here.
	// The handler would be w.handleVectorStoreAddMessage.
	go func() {
		// Simulated subscription loop
	}()

	<-ctx.Done()
	fmt.Println("QdrantWorker shutting down.")
	return nil
}

// handleVectorStoreAddMessage simulates processing an incoming NATS message for vector storage.
func (w *QdrantWorker) handleVectorStoreAddMessage(payload []byte) error {
	fmt.Printf("QdrantWorker received payload: %s\n", string(payload))

	var embeddingData EmbeddingData // Expecting EmbeddingData from EmbeddingWorker
	if err := json.Unmarshal(payload, &embeddingData); err != nil {
		fmt.Printf("QdrantWorker: Error unmarshalling EmbeddingData: %v\n", err)
		return fmt.Errorf("error unmarshalling EmbeddingData: %w", err)
	}
	fmt.Printf("QdrantWorker: Unmarshalled EmbeddingData for MemoryID: %s\n", embeddingData.MemoryID)

	if w.vs == nil {
		fmt.Println("QdrantWorker: VectorStore client is nil, cannot insert vectors.")
		return fmt.Errorf("VectorStore client is nil")
	}

	// Prepare VectorInput for VectorStore
	// Assuming Qdrant collection name is configured in w.cfg.VectorStoreConfig.Config.(*vectorstores.QdrantConfig).CollectionName
	collectionName := "default_collection" // Fallback
	if w.cfg.VectorStoreConfig != nil {
		if qdrantCfg, ok := w.cfg.VectorStoreConfig.Config.(*vectorstores.QdrantConfig); ok {
			collectionName = qdrantCfg.CollectionName
		} else {
			fmt.Println("QdrantWorker: VectorStoreConfig.Config is not *vectorstores.QdrantConfig")
		}
	} else {
		fmt.Println("QdrantWorker: VectorStoreConfig is nil in main Config")
	}

	vectorInput := vectorstores.VectorInput{
		ID:        embeddingData.MemoryID, // Using MemoryID as the vector ID
		Embedding: embeddingData.Embedding,
		Payload: map[string]interface{}{
			"text":          embeddingData.ProcessedText, // Or TextToEmbed
			"user_id":       embeddingData.UserID,
			"agent_id":      embeddingData.AgentID,
			"run_id":        embeddingData.RunID,
			"actor_id":      embeddingData.ActorID,                     // If ActorID was added to EmbeddingData from ProcessedMemoryData
			"original_text": embeddingData.TextToEmbed,                 // Assuming ProcessedText is the one embedded
			"timestamp":     time.Now().UTC().Format(time.RFC3339Nano), // Add a timestamp for the vector storage itself
			// Add any other relevant fields from embeddingData.BaseRequestInfo.Metadata
		},
	}
	if embeddingData.BaseRequestInfo.Metadata != nil {
		for k, v := range embeddingData.BaseRequestInfo.Metadata {
			vectorInput.Payload[k] = v
		}
	}

	fmt.Printf("QdrantWorker: Simulating VectorStore InsertVectors call for MemoryID: %s into collection %s\n", embeddingData.MemoryID, collectionName)
	err := w.vs.InsertVectors(collectionName, []vectorstores.VectorInput{vectorInput})
	if err != nil {
		fmt.Printf("QdrantWorker: Error simulating VectorStore InsertVectors: %v\n", err)
		return fmt.Errorf("error inserting vectors: %w", err)
	}
	fmt.Printf("QdrantWorker: Successfully simulated vector insertion for MemoryID: %s\n", embeddingData.MemoryID)

	// Simulate publishing MemoryEvent to TopicMemoryHistoryLog
	historyEvent := MemoryEvent{
		EventID:   uuid.New().String(),
		MemoryID:  embeddingData.MemoryID,
		EventType: "VECTOR_STORE_ADD",
		Timestamp: time.Now().UTC(),
		UserID:    embeddingData.UserID,
		AgentID:   embeddingData.AgentID,
		RunID:     embeddingData.RunID,
		ActorID:   embeddingData.ActorID,
		Details: map[string]interface{}{
			"collection_name": collectionName,
			"vector_id":       embeddingData.MemoryID,
			"embedding_dim":   len(embeddingData.Embedding),
		},
	}
	eventData, err := json.Marshal(historyEvent)
	if err != nil {
		fmt.Printf("QdrantWorker: Error marshalling MemoryEvent: %v\n", err)
	} else {
		if w.nc != nil {
			err = w.nc.Publish(context.Background(), w.cfg.TopicMemoryHistoryLog, eventData)
			if err != nil {
				fmt.Printf("QdrantWorker: Error publishing MemoryEvent to NATS topic %s: %v\n", w.cfg.TopicMemoryHistoryLog, err)
			} else {
				fmt.Printf("QdrantWorker: Published MemoryEvent to %s for MemoryID: %s\n", w.cfg.TopicMemoryHistoryLog, embeddingData.MemoryID)
			}
		} else {
			fmt.Printf("NATS_PUBLISH (QdrantWorker - nc is nil): Topic=%s, Payload=%s\n", w.cfg.TopicMemoryHistoryLog, string(eventData))
		}
	}

	return nil
}
