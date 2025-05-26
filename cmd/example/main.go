package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
 	"github.com/nats-io/nats.go"
 
 	"github.com/pnocera/gomem/pkg/memory"
 	"github.com/pnocera/gomem/pkg/natsclient" 
 )
 
 // NATSClientAdapter adapts the nats.Conn to the memory.NATSClient interface.
 type NATSClientAdapter struct {
 	nc *nats.Conn
 }
 
 func (a *NATSClientAdapter) Publish(ctx context.Context, topic string, data []byte) error {
 	// The natsclient.Publish doesn't currently use context, but we can add it if needed.
 	return natsclient.Publish(a.nc, topic, data)
 }
 
 func (a *NATSClientAdapter) Request(ctx context.Context, topic string, data []byte, timeout time.Duration) ([]byte, error) {
 	// The natsclient.Request uses its own context creation internally.
 	// If memory.NATSClient interface's context needs to be passed through,
 	// natsclient.Request would need modification. For now, we use its existing timeout.
 	msg, err := natsclient.Request(a.nc, topic, data, timeout)
 	if err != nil {
 		return nil, err
 	}
 	return msg.Data, nil
 }
 
 func (a *NATSClientAdapter) Subscribe(ctx context.Context, topic string, handler func(msg []byte)) error {
 	// The natsclient.Subscribe doesn't currently use context.
 	// The handler signature also differs (nats.MsgHandler vs func(msg []byte)).
 	// We'll wrap the handler.
 	_, err := natsclient.Subscribe(a.nc, topic, func(m *nats.Msg) {
 		handler(m.Data)
 	})
 	return err
 }
 
 func main() {
 	fmt.Println("--- Memory Package Integration Example with Real NATS Client ---")
 
 	// 1. memory.Config Unmarshalling and Validation
 	memoryConfigJSON := `{\n54| \t\t\"nats_address\": \"nats://localhost:4222\",\n55| \t\t\"openai_api_key\": \"sk-examplekey\",\n56| \t\t\"topic_memory_add_received\": \"mem0.memory.add.received\",\n57| \t\t\"topic_memory_process\": \"mem0.memory.process\",\n58| \t\t\"topic_memory_embed\": \"mem0.memory.embed\",\n59| \t\t\"topic_memory_vector_store_add\": \"mem0.memory.vectorstore.add\",\n60| \t\t\"topic_memory_graph_store_add\": \"mem0.memory.graphstore.add\",\n61| \t\t\"topic_memory_history_log\": \"mem0.memory.history.log\",\n62| \t\t\"topic_memory_search\": \"mem0.memory.search\",\n63| \t\t\"topic_memory_get\": \"mem0.memory.get\",\n64| \t\t\"topic_memory_update\": \"mem0.memory.update\",\n65| \t\t\"topic_memory_delete\": \"mem0.memory.delete\",\n66| \t\t\"enable_graph_store\": true,\n67| \t\t\"enable_infer\": true,\n68| \t\t\"graph_config\": {\n69| \t\t\t\"provider\": \"neo4j\",\n70| \t\t\t\"config\": {\n71| \t\t\t\t\"url\": \"bolt://graphdb:7687\",\n72| \t\t\t\t\"username\": \"neo4j\",\n73| \t\t\t\t\"password\": \"graphpassword\",\n74| \t\t\t\t\"database\": \"mem0graph\",\n75| \t\t\t\t\"base_label\": false\n76| \t\t\t},\n77| \t\t\t\"custom_prompt\": \"graph custom prompt for example\"\n78| \t\t},\n79| \t\t\"vector_store_config\": {\n80| \t\t\t\"provider\": \"qdrant\",\n81| \t\t\t\"config\": {\n82| \t\t\t\t\"address\": \"http://qdrant:6333\",\n83| \t\t\t\t\"collection_name\": \"mem0vectors\"\n84| \t\t\t}\n85| \t\t},\n86| \t\t\"custom_fact_extraction_prompt\": \"example custom fact extraction\",\n87| \t\t\"custom_update_memory_prompt\": \"example custom update memory\"\n88| \t}`
 
 	var memCfg memory.Config
 	err := json.Unmarshal([]byte(memoryConfigJSON), &memCfg)
 	if err != nil {
 		log.Fatalf("Error unmarshalling memory.Config: %v", err)
 	}
 
 	fmt.Printf("\\n--- Parsed memory.Config ---\\n")
 	fmt.Printf("NATS Address: %s\\n", memCfg.NATSAddress)
 	// ... (rest of the config printing logic remains the same)
 
 	// 2. Connect to NATS Server
 	log.Infof("Connecting to NATS server at %s...", memCfg.NATSAddress)
 	nc, err := natsclient.Connect(memCfg.NATSAddress)
 	if err != nil {
 		log.Fatalf("Error connecting to NATS: %v", err)
 	}
 	defer nc.Close()
 	log.Info("Successfully connected to NATS.")
 
 	// 3. Instantiate HistoryStore and MemoryService with the real NATS client
 	fmt.Printf("\\n--- MemoryService Operations ---\\n")
 	historyStore, err := memory.NewSQLiteHistoryStore(":memory:") // Use in-memory for example
 	if err != nil {
 		log.Fatalf("Error creating SQLiteHistoryStore: %v", err)
 	}
 	defer historyStore.Close()
 
 	natsAdapter := &NATSClientAdapter{nc: nc}
 	memoryService := memory.NewMemoryService(natsAdapter, &memCfg, historyStore)
 
 	// 4. Add Memory
 	addReq := memory.AddMemoryRequest{
 		BaseRequestInfo: memory.BaseRequestInfo{UserID: "example-user-123", AgentID: "example-agent-007"},
 		Messages: []memory.Message{
 			{Role: "user", Content: "Hello, this is a test memory sent via real NATS."},
 			{Role: "assistant", Content: "I acknowledge this test memory via real NATS."},
 		},
 		Infer: true,
 	}
 	fmt.Printf("\\nValidating AddMemoryRequest...\\n")
 	if err := addReq.Validate(); err != nil {
 		fmt.Printf("AddMemoryRequest validation error: %v\\n", err)
 	} else {
 		fmt.Println("AddMemoryRequest validated successfully.")
 	}
 
 	fmt.Printf("Calling memoryService.Add... (will publish to %s)\\n", memCfg.TopicMemoryAddReceived)
 	memoryID, err := memoryService.Add(context.Background(), &addReq)
 	if err != nil {
 		// This is expected if no worker is listening on TopicMemoryAddReceived to send a reply for the internal request.
 		// The Add method in MemoryService currently does a request internally.
 		fmt.Printf("memoryService.Add error: %v (This might be expected if no worker is processing the add request)\\n", err)
 	} else {
 		fmt.Printf("memoryService.Add success, MemoryID: %s\\n", memoryID)
 	}
 
 	// Example of subscribing to a topic (e.g., history log)
 	// This is just to demonstrate the Subscribe call.
 	// In a real application, workers would subscribe to relevant topics.
 	fmt.Printf("\\nAttempting to subscribe to topic: %s\\n", memCfg.TopicMemoryHistoryLog)
 	err = natsAdapter.Subscribe(context.Background(), memCfg.TopicMemoryHistoryLog, func(msg []byte) {
 		log.Infof("Received message on %s: %s", memCfg.TopicMemoryHistoryLog, string(msg))
 	})
 	if err != nil {
 		log.Errorf("Error subscribing to %s: %v", memCfg.TopicMemoryHistoryLog, err)
 	} else {
 		log.Infof("Successfully subscribed to %s. Listening for messages...", memCfg.TopicMemoryHistoryLog)
 	}
 
 	// 5. Search Memory
 	searchReq := memory.SearchMemoryRequest{
 		BaseRequestInfo: memory.BaseRequestInfo{UserID: "example-user-123"},
 		Query:           "test memory",
 		Limit:           5,
 	}
 	fmt.Printf("\\nValidating SearchMemoryRequest...\\n")
 	if err := searchReq.Validate(); err != nil {
 		fmt.Printf("SearchMemoryRequest validation error: %v\\n", err)
 	} else {
 		fmt.Println("SearchMemoryRequest validated successfully.")
 	}
 
 	fmt.Printf("Calling memoryService.Search... (will send request to %s)\\n", memCfg.TopicMemorySearch)
 	searchResults, err := memoryService.Search(context.Background(), &searchReq)
 	if err != nil {
 		// This is expected if no worker is listening on TopicMemorySearch
 		fmt.Printf("memoryService.Search error: %v (This is expected if no worker is processing search requests)\\n", err)
 	} else {
 		fmt.Printf("memoryService.Search results: %+v\\n", searchResults)
 	}
 
 	// 6. Get History
 	fmt.Printf("\\nCalling memoryService.GetHistory for dummy ID 'test-history-id'...\\n")
 	historyIDToFetch := "test-history-id"
 	if memoryID != "" { // memoryID might be empty if Add failed (e.g. no worker)
 		historyIDToFetch = memoryID
 		fmt.Printf("(Using actual memoryID from Add attempt: %s)\\n", memoryID)
 	}
 
 	historyEvents, err := memoryService.GetHistory(context.Background(), historyIDToFetch, memory.BaseRequestInfo{UserID: "example-user-123"})
 	if err != nil {
 		fmt.Printf("memoryService.GetHistory error: %v\\n", err)
 	} else {
 		fmt.Printf("memoryService.GetHistory success. Found %d events.\\n", len(historyEvents))
 		for i, event := range historyEvents {
 			fmt.Printf("  Event %d: ID=%s, Type=%s, Timestamp=%s\\n", i+1, event.EventID, event.EventType, event.Timestamp)
 		}
 	}
 
 	// Keep the main goroutine alive for a bit to see subscription messages if any
 	fmt.Println("\nExample finished. Sleeping for 5 seconds to allow potential NATS messages...")
 	time.Sleep(5 * time.Second)
 	log.Info("Exiting example.")
 }