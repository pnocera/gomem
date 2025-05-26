package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ProcessingWorker handles the initial processing of memories.
type ProcessingWorker struct {
	nc     NATSClient
	cfg    *Config
	openai OpenAIClient
}

// NewProcessingWorker creates a new ProcessingWorker.
func NewProcessingWorker(nc NATSClient, cfg *Config, openai OpenAIClient) *ProcessingWorker {
	return &ProcessingWorker{
		nc:     nc,
		cfg:    cfg,
		openai: openai,
	}
}

// Start begins the worker's NATS subscription.
func (w *ProcessingWorker) Start(ctx context.Context) error {
	if w.nc == nil {
		fmt.Println("ProcessingWorker: NATS client is nil, worker will not start.")
		// Block indefinitely or return an error, depending on desired behavior for nil NATS client
		<-ctx.Done()
		return nil // Or return an error indicating NATS client was not provided
	}

	fmt.Printf("ProcessingWorker started, listening on topic: %s\n", w.cfg.TopicMemoryProcess)
	// In a real implementation, w.nc.Subscribe would be called here.
	// The handler would be w.handleProcessMessage.
	// For shell, we simulate by just blocking.

	// Simulate a subscription loop that can be cancelled by the context
	go func() {
		// This is a simplified simulation. A real NATS subscription would handle this.
		// For now, we'll just print a message when the context is done.
		// To truly simulate receiving messages, we'd need a mock NATS client
		// or to integrate with a test NATS server.
		// For this subtask, handleProcessMessage will be called conceptually.
		// If there was an actual subscription:
		// err := w.nc.Subscribe(ctx, w.cfg.TopicMemoryProcess, w.handleProcessMessage)
		// if err != nil {
		//    log.Printf("ProcessingWorker: NATS subscription to %s failed: %v", w.cfg.TopicMemoryProcess, err)
		// }
	}()

	<-ctx.Done()
	fmt.Println("ProcessingWorker shutting down.")
	return nil
}

// handleProcessMessage simulates processing an incoming NATS message.
func (w *ProcessingWorker) handleProcessMessage(payload []byte) error {
	fmt.Printf("ProcessingWorker received payload: %s\n", string(payload))

	var addReq AddMemoryRequest // Assuming AddMemoryRequest is the input to this worker
	if err := json.Unmarshal(payload, &addReq); err != nil {
		fmt.Printf("ProcessingWorker: Error unmarshalling AddMemoryRequest: %v\n", err)
		return fmt.Errorf("error unmarshalling AddMemoryRequest: %w", err)
	}
	fmt.Printf("ProcessingWorker: Unmarshalled AddMemoryRequest for UserID: %s\n", addReq.UserID)

	// Simulate processing
	processedText := ""
	for _, msg := range addReq.Messages {
		processedText += msg.Content + " "
	}
	// Trim trailing space
	if len(processedText) > 0 {
		processedText = processedText[:len(processedText)-1]
	}

	memoryID := uuid.New().String() // Or use an ID from AddMemoryRequest if it were to carry one

	var extractedFacts []string
	if w.cfg.EnableInfer && w.openai != nil {
		fmt.Println("ProcessingWorker: Simulating OpenAI ExtractFacts call...")
		// In a real scenario, you'd pass appropriate text parts.
		// For shell, let's use the concatenated processedText.
		var textsToFactExtract []string
		for _, m := range addReq.Messages {
			textsToFactExtract = append(textsToFactExtract, m.Content)
		}

		factsString, err := w.openai.ExtractFacts(context.Background(), textsToFactExtract, w.cfg.CustomFactExtractionPrompt)
		if err != nil {
			fmt.Printf("ProcessingWorker: Error simulating OpenAI ExtractFacts: %v\n", err)
			// Decide if this is a fatal error or if processing can continue without facts
		} else {
			// Simulate splitting facts string into a slice
			extractedFacts = []string{factsString} // Simplified
			fmt.Printf("ProcessingWorker: Simulated extracted facts: %v\n", extractedFacts)
		}
	}

	processedData := ProcessedMemoryData{
		BaseRequestInfo:  addReq.BaseRequestInfo,
		OriginalMessages: addReq.Messages,
		ProcessedText:    processedText,
		MemoryID:         memoryID,
		ExtractedFacts:   extractedFacts,
	}

	jsonData, err := json.Marshal(processedData)
	if err != nil {
		fmt.Printf("ProcessingWorker: Error marshalling ProcessedMemoryData: %v\n", err)
		return fmt.Errorf("error marshalling ProcessedMemoryData: %w", err)
	}

	// Simulate publishing to TopicMemoryEmbed
	if w.nc != nil {
		err = w.nc.Publish(context.Background(), w.cfg.TopicMemoryEmbed, jsonData)
		if err != nil {
			fmt.Printf("ProcessingWorker: Error publishing to NATS topic %s: %v\n", w.cfg.TopicMemoryEmbed, err)
			// Potentially return error or implement retry
		} else {
			fmt.Printf("ProcessingWorker: Published ProcessedMemoryData to %s\n", w.cfg.TopicMemoryEmbed)
		}
	} else {
		fmt.Printf("NATS_PUBLISH (ProcessingWorker - nc is nil): Topic=%s, Payload=%s\n", w.cfg.TopicMemoryEmbed, string(jsonData))
	}

	// Simulate publishing MemoryEvent to TopicMemoryHistoryLog
	historyEvent := MemoryEvent{
		EventID:   uuid.New().String(),
		MemoryID:  memoryID,
		EventType: "MEMORY_PROCESSED",
		Timestamp: time.Now().UTC(),
		UserID:    addReq.UserID,
		AgentID:   addReq.AgentID,
		RunID:     addReq.RunID,
		ActorID:   addReq.ActorID,
		NewMemory: processedText, // Or a summary
		Details: map[string]interface{}{
			"original_message_count": len(addReq.Messages),
			"processed_text_length":  len(processedText),
			"facts_extracted_count":  len(extractedFacts),
		},
	}
	eventData, err := json.Marshal(historyEvent)
	if err != nil {
		fmt.Printf("ProcessingWorker: Error marshalling MemoryEvent: %v\n", err)
		// Log error, but typically don't let history logging failure stop main flow
	} else {
		if w.nc != nil {
			err = w.nc.Publish(context.Background(), w.cfg.TopicMemoryHistoryLog, eventData)
			if err != nil {
				fmt.Printf("ProcessingWorker: Error publishing MemoryEvent to NATS topic %s: %v\n", w.cfg.TopicMemoryHistoryLog, err)
			} else {
				fmt.Printf("ProcessingWorker: Published MemoryEvent to %s\n", w.cfg.TopicMemoryHistoryLog)
			}
		} else {
			fmt.Printf("NATS_PUBLISH (ProcessingWorker - nc is nil): Topic=%s, Payload=%s\n", w.cfg.TopicMemoryHistoryLog, string(eventData))
		}
	}

	return nil
}
