package memory

import (
	"context"
	"encoding/json"
	"fmt" // Assuming graphs config is needed for prompts
	"time"

	"github.com/pnocera/gomem/pkg/graphs"

	"github.com/google/uuid"
)

// DgraphWorker handles storing graph data in Dgraph (or a similar graph DB).
type DgraphWorker struct {
	nc       NATSClient
	cfg      *Config
	openai   OpenAIClient
	dg       DgraphClient             // Dgraph client interface
	graphCfg *graphs.GraphStoreConfig // For graph-specific prompts or settings
}

// NewDgraphWorker creates a new DgraphWorker.
func NewDgraphWorker(nc NATSClient, cfg *Config, openai OpenAIClient, dg DgraphClient, graphCfg *graphs.GraphStoreConfig) *DgraphWorker {
	return &DgraphWorker{
		nc:       nc,
		cfg:      cfg,
		openai:   openai,
		dg:       dg,
		graphCfg: graphCfg,
	}
}

// Start begins the worker's NATS subscription.
func (w *DgraphWorker) Start(ctx context.Context) error {
	if !w.cfg.EnableGraphStore {
		fmt.Println("DgraphWorker: Graph store is disabled in config, worker will not start.")
		// Block or return based on desired behavior when disabled
		<-ctx.Done()
		return nil
	}
	if w.nc == nil {
		fmt.Println("DgraphWorker: NATS client is nil, worker will not start.")
		<-ctx.Done()
		return nil
	}
	if w.dg == nil {
		fmt.Println("DgraphWorker: Dgraph client (dg) is nil, worker will not start effectively.")
	}
	if w.openai == nil {
		fmt.Println("DgraphWorker: OpenAI client is nil, graph data extraction will be skipped.")
	}

	fmt.Printf("DgraphWorker started, listening on topic: %s\n", w.cfg.TopicMemoryGraphStoreAdd)
	// In a real implementation, w.nc.Subscribe would be called here.
	// The handler would be w.handleGraphStoreAddMessage.
	go func() {
		// Simulated subscription loop
	}()

	<-ctx.Done()
	fmt.Println("DgraphWorker shutting down.")
	return nil
}

// handleGraphStoreAddMessage simulates processing an incoming NATS message for graph storage.
func (w *DgraphWorker) handleGraphStoreAddMessage(payload []byte) error {
	fmt.Printf("DgraphWorker received payload: %s\n", string(payload))

	var graphData GraphStoreStorageData // Expecting GraphStoreStorageData
	if err := json.Unmarshal(payload, &graphData); err != nil {
		fmt.Printf("DgraphWorker: Error unmarshalling GraphStoreStorageData: %v\n", err)
		return fmt.Errorf("error unmarshalling GraphStoreStorageData: %w", err)
	}
	fmt.Printf("DgraphWorker: Unmarshalled GraphStoreStorageData for MemoryID: %s\n", graphData.MemoryID)

	if w.dg == nil {
		fmt.Println("DgraphWorker: Dgraph client is nil, cannot store graph data.")
		return fmt.Errorf("Dgraph client is nil")
	}

	// Simulate OpenAI ExtractGraphData if not already populated and OpenAI client exists
	if (len(graphData.Entities) == 0 || len(graphData.Relationships) == 0) && w.openai != nil {
		fmt.Println("DgraphWorker: Simulating OpenAI ExtractGraphData call...")
		customPrompt := ""
		if w.graphCfg != nil {
			customPrompt = w.graphCfg.CustomPrompt
		} else if w.cfg.CustomFactExtractionPrompt != "" { // Fallback to general fact extraction prompt if specific graph prompt not set
			customPrompt = w.cfg.CustomFactExtractionPrompt
		}

		entities, relations, err := w.openai.ExtractGraphData(context.Background(), graphData.TextForGraph, customPrompt)
		if err != nil {
			fmt.Printf("DgraphWorker: Error simulating OpenAI ExtractGraphData: %v\n", err)
			// Decide if this is fatal or proceed without graph data
		} else {
			graphData.Entities = entities
			graphData.Relationships = relations
			fmt.Printf("DgraphWorker: Simulated graph data extraction for MemoryID: %s. Entities: %d, Relations: %d\n", graphData.MemoryID, len(entities), len(relations))
		}
	} else if w.openai == nil {
		fmt.Println("DgraphWorker: OpenAI client is nil, skipping graph data extraction by OpenAI.")
	}

	if len(graphData.Entities) > 0 || len(graphData.Relationships) > 0 {
		fmt.Printf("DgraphWorker: Simulating Dgraph Mutate call for MemoryID: %s\n", graphData.MemoryID)
		// In a real scenario, you'd transform graphData.Entities and graphData.Relationships
		// into the format expected by dg.Mutate.
		// For shell, we can just pass the struct, or a simplified map.
		mockMutationData := map[string]interface{}{
			"memoryId":      graphData.MemoryID,
			"entities":      graphData.Entities,
			"relationships": graphData.Relationships,
		}
		err := w.dg.Mutate(context.Background(), mockMutationData)
		if err != nil {
			fmt.Printf("DgraphWorker: Error simulating Dgraph Mutate: %v\n", err)
			return fmt.Errorf("error mutating graph data: %w", err)
		}
		fmt.Printf("DgraphWorker: Successfully simulated graph data mutation for MemoryID: %s\n", graphData.MemoryID)
	} else {
		fmt.Printf("DgraphWorker: No entities or relationships to store for MemoryID: %s\n", graphData.MemoryID)
	}

	// Simulate publishing MemoryEvent to TopicMemoryHistoryLog
	historyEvent := MemoryEvent{
		EventID:   uuid.New().String(),
		MemoryID:  graphData.MemoryID,
		EventType: "GRAPH_STORE_ADD",
		Timestamp: time.Now().UTC(),
		UserID:    graphData.UserID,
		AgentID:   graphData.AgentID,
		RunID:     graphData.RunID,
		ActorID:   graphData.ActorID,
		Details: map[string]interface{}{
			"entities_count":      len(graphData.Entities),
			"relationships_count": len(graphData.Relationships),
		},
	}
	eventData, err := json.Marshal(historyEvent)
	if err != nil {
		fmt.Printf("DgraphWorker: Error marshalling MemoryEvent: %v\n", err)
	} else {
		if w.nc != nil {
			err = w.nc.Publish(context.Background(), w.cfg.TopicMemoryHistoryLog, eventData)
			if err != nil {
				fmt.Printf("DgraphWorker: Error publishing MemoryEvent to NATS topic %s: %v\n", w.cfg.TopicMemoryHistoryLog, err)
			} else {
				fmt.Printf("DgraphWorker: Published MemoryEvent to %s for MemoryID: %s\n", w.cfg.TopicMemoryHistoryLog, graphData.MemoryID)
			}
		} else {
			fmt.Printf("NATS_PUBLISH (DgraphWorker - nc is nil): Topic=%s, Payload=%s\n", w.cfg.TopicMemoryHistoryLog, string(eventData))
		}
	}

	return nil
}
