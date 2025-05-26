package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"time"

	"github.com/pnocera/gomem/pkg/graphs"
	"github.com/pnocera/gomem/pkg/memory"
	"github.com/pnocera/gomem/pkg/vectorstores"
	
)

// Mock NATSClient for example usage
type mockNATSClient struct{}

func (m *mockNATSClient) Publish(ctx context.Context, topic string, data []byte) error {
	fmt.Printf("MOCK_NATS_PUBLISH: Topic=%s, Data=%s\n", topic, string(data))
	return nil
}

func (m *mockNATSClient) Request(ctx context.Context, topic string, data []byte, timeout time.Duration) ([]byte, error) {
	fmt.Printf("MOCK_NATS_REQUEST: Topic=%s, Data=%s\n", topic, string(data))
	// For Search and Get, the shell service expects an error to indicate "not fully implemented"
	// For this example, we'll return a generic error for Request.
	// A more sophisticated mock might return specific JSON based on the topic.
	// The actual topic names will come from the memCfg passed to the service.
	if topic == "mem0.memory.search" { // Example topic, replace with actual from memCfg if needed for specific mock logic
		return []byte("[]"), fmt.Errorf("mockNATSClient: Search via NATS not fully implemented (response handling pending)")
	}
	if topic == "mem0.memory.get" { // Example topic
		return nil, fmt.Errorf("mockNATSClient: Get via NATS not fully implemented (response handling pending)")
	}
	return nil, fmt.Errorf("mockNATSClient: Request not fully implemented for topic %s", topic)
}

func (m *mockNATSClient) Subscribe(ctx context.Context, topic string, handler func(msg []byte)) error {
	fmt.Printf("MOCK_NATS_SUBSCRIBE: Topic=%s\n", topic)
	// In a real mock for testing workers, you might launch a goroutine to send mock messages to the handler.
	// For this example, doing nothing is fine as service methods don't use Subscribe directly.
	return nil
}

func main() {
	fmt.Println("--- Memory Package Integration Example ---")

	// 1. memory.Config Unmarshalling and Validation
	memoryConfigJSON := `{
		"nats_address": "nats://localhost:4222",
		"openai_api_key": "sk-examplekey",
		"topic_memory_add_received": "mem0.memory.add.received",
		"topic_memory_process": "mem0.memory.process",
		"topic_memory_embed": "mem0.memory.embed",
		"topic_memory_vector_store_add": "mem0.memory.vectorstore.add",
		"topic_memory_graph_store_add": "mem0.memory.graphstore.add",
		"topic_memory_history_log": "mem0.memory.history.log",
		"topic_memory_search": "mem0.memory.search",
		"topic_memory_get": "mem0.memory.get",
		"topic_memory_update": "mem0.memory.update",
		"topic_memory_delete": "mem0.memory.delete",
		"enable_graph_store": true,
		"enable_infer": true,
		"graph_config": {
			"provider": "neo4j",
			"config": {
				"url": "bolt://graphdb:7687",
				"username": "neo4j",
				"password": "graphpassword",
				"database": "mem0graph",
				"base_label": false
			},
			"custom_prompt": "graph custom prompt for example"
		},
		"vector_store_config": {
			"provider": "qdrant",
			"config": {
				"address": "http://qdrant:6333",
				"collection_name": "mem0vectors"
			}
		},
		"custom_fact_extraction_prompt": "example custom fact extraction",
		"custom_update_memory_prompt": "example custom update memory"
	}`

	var memCfg memory.Config
	err := json.Unmarshal([]byte(memoryConfigJSON), &memCfg)
	if err != nil {
		log.Fatalf("Error unmarshalling memory.Config: %v", err)
	}

	fmt.Printf("\n--- Parsed memory.Config ---\n")
	fmt.Printf("NATS Address: %s\n", memCfg.NATSAddress)
	fmt.Printf("EnableGraphStore: %t\n", memCfg.EnableGraphStore)

	if memCfg.GraphConfig != nil {
		fmt.Printf("GraphConfig Provider: %s\n", memCfg.GraphConfig.Provider)
		if neo4jConf, ok := memCfg.GraphConfig.Config.(*graphs.Neo4jConfig); ok {
			fmt.Printf("  Neo4j URL: %s, Database: %s\n", neo4jConf.URL, neo4jConf.Database)
		} else {
			fmt.Printf("  Could not assert GraphConfig.Config to *graphs.Neo4jConfig (Type: %T)\n", memCfg.GraphConfig.Config)
		}
	} else {
		fmt.Println("GraphConfig is nil")
	}

	if memCfg.VectorStoreConfig != nil {
		fmt.Printf("VectorStoreConfig Provider: %s\n", memCfg.VectorStoreConfig.Provider)
		if qdrantConf, ok := memCfg.VectorStoreConfig.Config.(*vectorstores.QdrantConfig); ok {
			fmt.Printf("  Qdrant Address: %s, Collection: %s\n", qdrantConf.Address, qdrantConf.CollectionName)
		} else {
			fmt.Printf("  Could not assert VectorStoreConfig.Config to *vectorstores.QdrantConfig (Type: %T)\n", memCfg.VectorStoreConfig.Config)
		}
	} else {
		fmt.Println("VectorStoreConfig is nil")
	}

	fmt.Printf("\nValidating memory.Config...\n")
	err = memCfg.Validate()
	if err != nil {
		fmt.Printf("memory.Config validation error: %v\n", err)
	} else {
		fmt.Println("memory.Config validated successfully.")
	}

	// 2. Instantiate HistoryStore and MemoryService
	fmt.Printf("\n--- MemoryService Operations ---\n")
	historyStore, err := memory.NewSQLiteHistoryStore(":memory:") // Use in-memory for example
	if err != nil {
		log.Fatalf("Error creating SQLiteHistoryStore: %v", err)
	}
	defer historyStore.Close()

	mockNatsClient := &mockNATSClient{}
	memoryService := memory.NewMemoryService(mockNatsClient, &memCfg, historyStore)

	// 3. Add Memory
	addReq := memory.AddMemoryRequest{
		BaseRequestInfo: memory.BaseRequestInfo{UserID: "example-user-123", AgentID: "example-agent-007"},
		Messages: []memory.Message{
			{Role: "user", Content: "Hello, this is a test memory."},
			{Role: "assistant", Content: "I acknowledge this test memory."},
		},
		Infer: true,
	}
	fmt.Printf("\nValidating AddMemoryRequest...\n")
	if err := addReq.Validate(); err != nil { // Assuming Validate method exists
		fmt.Printf("AddMemoryRequest validation error: %v\n", err)
	} else {
		fmt.Println("AddMemoryRequest validated successfully.")
	}

	fmt.Printf("Calling memoryService.Add...\n")
	memoryID, err := memoryService.Add(context.Background(), &addReq)
	if err != nil {
		fmt.Printf("memoryService.Add error: %v\n", err)
	} else {
		fmt.Printf("memoryService.Add success, MemoryID: %s\n", memoryID)
	}

	// 4. Search Memory
	searchReq := memory.SearchMemoryRequest{
		BaseRequestInfo: memory.BaseRequestInfo{UserID: "example-user-123"},
		Query:           "test memory",
		Limit:           5,
	}
	fmt.Printf("\nValidating SearchMemoryRequest...\n")
	if err := searchReq.Validate(); err != nil { // Assuming Validate method exists
		fmt.Printf("SearchMemoryRequest validation error: %v\n", err)
	} else {
		fmt.Println("SearchMemoryRequest validated successfully.")
	}

	fmt.Printf("Calling memoryService.Search...\n")
	searchResults, err := memoryService.Search(context.Background(), &searchReq)
	if err != nil {
		fmt.Printf("memoryService.Search error: %v\n", err) // Expected from shell service
	}
	fmt.Printf("memoryService.Search results (shell): %+v\n", searchResults)

	// 5. Get History
	fmt.Printf("\nCalling memoryService.GetHistory for dummy ID 'test-history-id'...\n")
	// Use the actual memoryID from the Add operation if successful for a more meaningful history
	historyIDToFetch := "test-history-id"
	if memoryID != "" {
		historyIDToFetch = memoryID
		fmt.Printf("(Using actual memoryID from Add: %s)\n", memoryID)
	}

	historyEvents, err := memoryService.GetHistory(context.Background(), historyIDToFetch, memory.BaseRequestInfo{UserID: "example-user-123"})
	if err != nil {
		fmt.Printf("memoryService.GetHistory error: %v\n", err)
	} else {
		fmt.Printf("memoryService.GetHistory success. Found %d events.\n", len(historyEvents))
		for i, event := range historyEvents {
			fmt.Printf("  Event %d: ID=%s, Type=%s, Timestamp=%s\n", i+1, event.EventID, event.EventType, event.Timestamp)
		}
		if len(historyEvents) == 0 && historyIDToFetch == memoryID {
			fmt.Println("  (Note: History is empty for the added memory because workers are not actually running to log events like MEMORY_PROCESSED etc.)")
		}
	}
}
