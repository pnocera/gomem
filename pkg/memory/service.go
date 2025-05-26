package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MemoryService defines the interface for high-level memory operations.
type MemoryService interface {
	Add(ctx context.Context, req *AddMemoryRequest) (memoryID string, err error)
	Search(ctx context.Context, req *SearchMemoryRequest) ([]MemoryResult, error)
	Get(ctx context.Context, memoryID string, baseInfo BaseRequestInfo) (*MemoryResult, error)
	Update(ctx context.Context, memoryID string, data map[string]interface{}, baseInfo BaseRequestInfo) error
	Delete(ctx context.Context, memoryID string, baseInfo BaseRequestInfo) error
	GetHistory(ctx context.Context, memoryID string, baseInfo BaseRequestInfo) ([]*MemoryEvent, error)
}

// memoryServiceImpl implements the MemoryService interface.
type memoryServiceImpl struct {
	nc      NATSClient
	cfg     *Config
	history HistoryStore
	// openai OpenAIClient // Placeholder
}

// Compile-time check to ensure *memoryServiceImpl satisfies the MemoryService interface.
var _ MemoryService = (*memoryServiceImpl)(nil)

// NewMemoryService creates a new instance of memoryServiceImpl.
func NewMemoryService(nc NATSClient, cfg *Config, historyStore HistoryStore) MemoryService {
	return &memoryServiceImpl{
		nc:      nc,
		cfg:     cfg,
		history: historyStore,
	}
}

// Add handles adding a new memory.
func (s *memoryServiceImpl) Add(ctx context.Context, req *AddMemoryRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", fmt.Errorf("invalid AddMemoryRequest: %w", err)
	}

	memoryID := uuid.New().String()
	// Conceptual: Populate parts of ProcessedMemoryData if needed before publishing
	// For this shell, we'll assume the AddMemoryRequest itself is the payload,
	// or a simple derivative. The prompt mentions publishing AddMemoryRequest or ProcessedMemoryData.
	// Let's use AddMemoryRequest for simplicity in the NATS message for now.
	// A real system might have an initial processing step before this first publish.

	// Assign the generated memoryID to the request for downstream consumers if it's part of the NATS message.
	// However, AddMemoryRequest doesn't have a memoryID field.
	// Let's define a payload struct for NATS if AddMemoryRequest isn't directly used.
	// For this example, we'll marshal 'req' and assume downstream services handle ID generation if needed
	// or use a wrapper. The requirement is to publish to TopicMemoryAddReceived.

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal AddMemoryRequest: %w", err)
	}

	if s.nc != nil {
		err = s.nc.Publish(ctx, s.cfg.TopicMemoryAddReceived, jsonData)
		if err != nil {
			return "", fmt.Errorf("failed to publish to NATS topic %s: %w", s.cfg.TopicMemoryAddReceived, err)
		}
	} else {
		fmt.Printf("NATS_PUBLISH (nc is nil): Topic=%s, Payload=%s\n", s.cfg.TopicMemoryAddReceived, string(jsonData))
	}

	return memoryID, nil
}

// Search handles searching memories.
func (s *memoryServiceImpl) Search(ctx context.Context, req *SearchMemoryRequest) ([]MemoryResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SearchMemoryRequest: %w", err)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SearchMemoryRequest: %w", err)
	}

	if s.nc != nil {
		// Define a reasonable timeout for NATS request-reply
		timeout := 5 * time.Second // Example timeout
		_, err = s.nc.Request(ctx, s.cfg.TopicMemorySearch, jsonData, timeout)
		if err != nil {
			return nil, fmt.Errorf("NATS request to %s failed: %w", s.cfg.TopicMemorySearch, err)
		}
		// TODO: Unmarshal responseData into []MemoryResult
		return nil, fmt.Errorf("Search via NATS not fully implemented (response handling pending)")
	}

	fmt.Printf("NATS_REQUEST (nc is nil): Topic=%s, Payload=%s\n", s.cfg.TopicMemorySearch, string(jsonData))
	return nil, fmt.Errorf("Search via NATS not fully implemented (NATS client is nil)")
}

// GetRequestData is a helper struct for Get, Update, Delete operations
type GetRequestData struct {
	BaseRequestInfo
	MemoryID string `json:"memory_id"`
}

// UpdateRequestData is a helper struct for Update operation
type UpdateRequestData struct {
	BaseRequestInfo
	MemoryID string                 `json:"memory_id"`
	Data     map[string]interface{} `json:"data"`
}

// Get retrieves a specific memory.
func (s *memoryServiceImpl) Get(ctx context.Context, memoryID string, baseInfo BaseRequestInfo) (*MemoryResult, error) {
	if memoryID == "" {
		return nil, fmt.Errorf("memoryID cannot be empty")
	}

	payload := GetRequestData{
		MemoryID:        memoryID,
		BaseRequestInfo: baseInfo,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Get request: %w", err)
	}

	if s.nc != nil {
		timeout := 5 * time.Second // Example timeout
		_, err = s.nc.Request(ctx, s.cfg.TopicMemoryGet, jsonData, timeout)
		if err != nil {
			return nil, fmt.Errorf("NATS request to %s failed: %w", s.cfg.TopicMemoryGet, err)
		}
		// TODO: Unmarshal responseData into *MemoryResult
		return nil, fmt.Errorf("Get via NATS not fully implemented (response handling pending)")
	}

	fmt.Printf("NATS_REQUEST (nc is nil): Topic=%s, Payload=%s\n", s.cfg.TopicMemoryGet, string(jsonData))
	return nil, fmt.Errorf("Get via NATS not fully implemented (NATS client is nil)")
}

// Update updates a specific memory.
// Note: The prompt mentions a conceptual `s.cfg.TopicMemoryUpdate`. This needs to be added to `Config` struct.
func (s *memoryServiceImpl) Update(ctx context.Context, memoryID string, data map[string]interface{}, baseInfo BaseRequestInfo) error {
	if memoryID == "" {
		return fmt.Errorf("memoryID cannot be empty")
	}

	payload := UpdateRequestData{
		MemoryID:        memoryID,
		Data:            data,
		BaseRequestInfo: baseInfo,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Update request: %w", err)
	}

	// Assuming TopicMemoryUpdate will be added to s.cfg
	// For now, using a placeholder string if it's not in Config.
	topic := "mem0.memory.update"                      // Placeholder
	if s.cfg != nil && s.cfg.TopicMemoryUpdate != "" { // Check if TopicMemoryUpdate is defined
		topic = s.cfg.TopicMemoryUpdate
	}

	if s.nc != nil {
		err = s.nc.Publish(ctx, topic, jsonData)
		if err != nil {
			return fmt.Errorf("failed to publish to NATS topic %s: %w", topic, err)
		}
		return nil // Or handle response if it's a request-reply
	}

	fmt.Printf("NATS_PUBLISH (nc is nil): Topic=%s, Payload=%s\n", topic, string(jsonData))
	return fmt.Errorf("Update via NATS not fully implemented (NATS client is nil)")
}

// Delete removes a specific memory.
// Note: The prompt mentions a conceptual `s.cfg.TopicMemoryDelete`. This needs to be added to `Config` struct.
func (s *memoryServiceImpl) Delete(ctx context.Context, memoryID string, baseInfo BaseRequestInfo) error {
	if memoryID == "" {
		return fmt.Errorf("memoryID cannot be empty")
	}

	payload := GetRequestData{ // Using GetRequestData as it fits the payload needs (MemoryID + BaseInfo)
		MemoryID:        memoryID,
		BaseRequestInfo: baseInfo,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Delete request: %w", err)
	}

	topic := "mem0.memory.delete"                      // Placeholder
	if s.cfg != nil && s.cfg.TopicMemoryDelete != "" { // Check if TopicMemoryDelete is defined
		topic = s.cfg.TopicMemoryDelete
	}

	if s.nc != nil {
		err = s.nc.Publish(ctx, topic, jsonData)
		if err != nil {
			return fmt.Errorf("failed to publish to NATS topic %s: %w", topic, err)
		}
		return nil // Or handle response if it's a request-reply
	}

	fmt.Printf("NATS_PUBLISH (nc is nil): Topic=%s, Payload=%s\n", topic, string(jsonData))
	return fmt.Errorf("Delete via NATS not fully implemented (NATS client is nil)")
}

// GetHistory retrieves memory events directly from the history store.
func (s *memoryServiceImpl) GetHistory(ctx context.Context, memoryID string, baseInfo BaseRequestInfo) ([]*MemoryEvent, error) {
	// Here baseInfo might be used for authorization/filtering in a more complex setup,
	// but the current HistoryStore interface doesn't use it for GetHistory.
	if memoryID == "" {
		return nil, fmt.Errorf("memoryID cannot be empty")
	}
	if s.history == nil {
		return nil, fmt.Errorf("history store is not initialized")
	}
	// The current HistoryStore GetHistory only takes memoryID.
	// If BaseRequestInfo is needed for filtering at history store level, the interface would need an update.
	// For now, we pass only memoryID.
	events, err := s.history.GetHistory(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history from store: %w", err)
	}
	// The interface expects []*MemoryEvent, which s.history.GetHistory returns.
	return events, err
}
