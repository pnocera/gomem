package memory

import (
	"context"
	"encoding/json"
	"fmt"
)

// HistoryWorker handles logging memory events from NATS.
type HistoryWorker struct {
	nc           NATSClient
	cfg          *Config
	historyStore HistoryStore
}

// NewHistoryWorker creates a new HistoryWorker.
func NewHistoryWorker(nc NATSClient, cfg *Config, historyStore HistoryStore) *HistoryWorker {
	return &HistoryWorker{
		nc:           nc,
		cfg:          cfg,
		historyStore: historyStore,
	}
}

// Start begins the worker's NATS subscription.
func (w *HistoryWorker) Start(ctx context.Context) error {
	if w.nc == nil {
		fmt.Println("HistoryWorker: NATS client is nil, worker will not start.")
		<-ctx.Done()
		return nil
	}
	if w.historyStore == nil {
		fmt.Println("HistoryWorker: HistoryStore is nil, worker will not start effectively.")
	}

	fmt.Printf("HistoryWorker started, listening on topic: %s\n", w.cfg.TopicMemoryHistoryLog)
	// In a real implementation, w.nc.Subscribe would be called here.
	// The handler would be w.handleHistoryLogMessage.
	go func() {
		// Simulated subscription loop
	}()

	<-ctx.Done()
	fmt.Println("HistoryWorker shutting down.")
	return nil
}

// handleHistoryLogMessage simulates processing an incoming NATS message for history logging.
func (w *HistoryWorker) handleHistoryLogMessage(payload []byte) error {
	fmt.Printf("HistoryWorker received payload: %s\n", string(payload))

	var event MemoryEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		fmt.Printf("HistoryWorker: Error unmarshalling MemoryEvent: %v\n", err)
		return fmt.Errorf("error unmarshalling MemoryEvent: %w", err)
	}
	fmt.Printf("HistoryWorker: Unmarshalled MemoryEvent ID: %s, Type: %s\n", event.EventID, event.EventType)

	if w.historyStore == nil {
		fmt.Println("HistoryWorker: HistoryStore is nil, cannot log event.")
		return fmt.Errorf("HistoryStore is nil")
	}

	fmt.Printf("HistoryWorker: Simulating HistoryStore LogEvent call for EventID: %s\n", event.EventID)
	err := w.historyStore.LogEvent(context.Background(), &event) // Pass context if needed by actual store
	if err != nil {
		fmt.Printf("HistoryWorker: Error simulating HistoryStore LogEvent: %v\n", err)
		return fmt.Errorf("error logging event to history store: %w", err)
	}
	fmt.Printf("HistoryWorker: Successfully simulated logging event for EventID: %s\n", event.EventID)

	return nil
}
