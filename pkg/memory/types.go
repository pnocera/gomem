package memory

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// BaseRequestInfo contains common fields for requests.
type BaseRequestInfo struct {
	UserID   string                 `json:"user_id,omitempty"`
	AgentID  string                 `json:"agent_id,omitempty"`
	RunID    string                 `json:"run_id,omitempty"`
	ActorID  string                 `json:"actor_id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role" validate:"required,oneof=user assistant system"`
	Content string `json:"content" validate:"required"`
	Name    string `json:"name,omitempty"` // For actor_id in messages
}

// Validate validates the Message struct.
func (m *Message) Validate() error {
	validate := validator.New()
	return validate.Struct(m)
}

// AddMemoryRequest is the payload for adding a new memory.
type AddMemoryRequest struct {
	BaseRequestInfo
	Messages   []Message `json:"messages" validate:"required,min=1,dive"` // dive validates each element in slice
	Infer      bool      `json:"infer"`
	MemoryType string    `json:"memory_type,omitempty"`
	Prompt     string    `json:"prompt,omitempty"`
}

// Validate validates the AddMemoryRequest struct.
func (r *AddMemoryRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ProcessedMemoryData is the data after initial LLM processing.
type ProcessedMemoryData struct {
	BaseRequestInfo
	OriginalMessages []Message `json:"original_messages"`
	ProcessedText    string    `json:"processed_text"`
	MemoryID         string    `json:"memory_id"`
	ExtractedFacts   []string  `json:"extracted_facts,omitempty"`
}

// EmbeddingData contains text and its embedding.
type EmbeddingData struct {
	BaseRequestInfo
	MemoryID      string    `json:"memory_id"`
	TextToEmbed   string    `json:"text_to_embed"`
	Embedding     []float32 `json:"embedding"`
	ProcessedText string    `json:"processed_text"`
}

// VectorStoreStorageData is for the Qdrant worker.
type VectorStoreStorageData struct {
	BaseRequestInfo
	MemoryID  string    `json:"memory_id"`
	Embedding []float32 `json:"embedding"`
	Text      string    `json:"text"`
	Role      string    `json:"role,omitempty"`
	ActorID   string    `json:"actor_id,omitempty"` // Explicitly from message if available
	Timestamp time.Time `json:"timestamp"`
	// Other metadata from BaseRequestInfo.Metadata will be part of Payload in VectorInput
}

// Entity for GraphStoreStorageData - minimal for now.
type Entity struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// Relation for GraphStoreStorageData - minimal for now.
type Relation struct {
	SourceID         string `json:"source_id"`
	TargetID         string `json:"target_id"`
	RelationshipType string `json:"relationship_type"`
}

// GraphStoreStorageData is for the Dgraph worker.
type GraphStoreStorageData struct {
	BaseRequestInfo
	MemoryID      string     `json:"memory_id"`
	TextForGraph  string     `json:"text_for_graph"`
	Entities      []Entity   `json:"entities,omitempty"`
	Relationships []Relation `json:"relationships,omitempty"`
}

// MemoryEvent for history logging.
type MemoryEvent struct {
	EventID     string                 `json:"event_id"`
	MemoryID    string                 `json:"memory_id,omitempty"`
	EventType   string                 `json:"event_type" validate:"required"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id,omitempty"`  // Duplicates BaseRequestInfo but can be explicit for event log
	AgentID     string                 `json:"agent_id,omitempty"` // Duplicates BaseRequestInfo
	RunID       string                 `json:"run_id,omitempty"`   // Duplicates BaseRequestInfo
	ActorID     string                 `json:"actor_id,omitempty"` // Duplicates BaseRequestInfo
	OldMemory   string                 `json:"old_memory,omitempty"`
	NewMemory   string                 `json:"new_memory,omitempty"`
	SearchQuery string                 `json:"search_query,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// Validate validates the MemoryEvent struct.
func (e *MemoryEvent) Validate() error {
	validate := validator.New()
	return validate.Struct(e)
}

// GraphRelation for MemoryResult - minimal for now.
type GraphRelation struct {
	SourceNodeID string `json:"source_node_id"`
	TargetNodeID string `json:"target_node_id"`
	Type         string `json:"type"`
	// Add other properties if needed
}

// MemoryResult is the structure for returning memories.
type MemoryResult struct {
	ID        string                 `json:"id"`
	Memory    string                 `json:"memory"`
	Score     float32                `json:"score,omitempty"`
	Hash      string                 `json:"hash,omitempty"`
	CreatedAt time.Time              `json:"created_at,omitempty"`
	UpdatedAt time.Time              `json:"updated_at,omitempty"`
	UserID    string                 `json:"user_id,omitempty"` // Explicit fields for common query/filter needs
	AgentID   string                 `json:"agent_id,omitempty"`
	RunID     string                 `json:"run_id,omitempty"`
	ActorID   string                 `json:"actor_id,omitempty"`
	Role      string                 `json:"role,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Relations []GraphRelation        `json:"relations,omitempty"`
}

// SearchMemoryRequest
type SearchMemoryRequest struct {
	BaseRequestInfo
	Query string `json:"query" validate:"required"`
	Limit int    `json:"limit" validate:"omitempty,gt=0"` // Default handling (e.g., 100) done in processing logic
}

// Validate validates the SearchMemoryRequest struct.
func (r *SearchMemoryRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
