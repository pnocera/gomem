package memory

import (
	"fmt"

	"github.com/pnocera/gomem/pkg/graphs"
	"github.com/pnocera/gomem/pkg/vectorstores"

	"github.com/go-playground/validator/v10"
)

// Config holds configuration for the memory service.
type Config struct {
	NATSAddress  string `json:"nats_address" validate:"required"`
	OpenAIAPIKey string `json:"openai_api_key" validate:"required"`

	// NATS Topics
	TopicMemoryAddReceived    string `json:"topic_memory_add_received" validate:"required"`
	TopicMemoryProcess        string `json:"topic_memory_process" validate:"required"`
	TopicMemoryEmbed          string `json:"topic_memory_embed" validate:"required"`
	TopicMemoryVectorStoreAdd string `json:"topic_memory_vector_store_add" validate:"required"`
	TopicMemoryGraphStoreAdd  string `json:"topic_memory_graph_store_add" validate:"required"`
	TopicMemoryHistoryLog     string `json:"topic_memory_history_log" validate:"required"`
	TopicMemorySearch         string `json:"topic_memory_search" validate:"required"`
	TopicMemoryGet            string `json:"topic_memory_get" validate:"required"`
	TopicMemoryUpdate         string `json:"topic_memory_update" validate:"required"`
	TopicMemoryDelete         string `json:"topic_memory_delete" validate:"required"`

	// Feature flags
	EnableGraphStore bool `json:"enable_graph_store"`
	EnableInfer      bool `json:"enable_infer"` // default:"true" is conceptual, Go uses zero value (false)

	GraphConfig       *graphs.GraphStoreConfig        `json:"graph_config,omitempty"`
	VectorStoreConfig *vectorstores.VectorStoreConfig `json:"vector_store_config,omitempty"`

	CustomFactExtractionPrompt string `json:"custom_fact_extraction_prompt,omitempty"`
	CustomUpdateMemoryPrompt   string `json:"custom_update_memory_prompt,omitempty"`
}

// Validate validates the Config struct.
func (c *Config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}

	if c.GraphConfig != nil {
		if err := c.GraphConfig.Validate(); err != nil {
			return fmt.Errorf("graph_config validation failed: %w", err)
		}
	}
	if c.VectorStoreConfig != nil {
		if err := c.VectorStoreConfig.Validate(); err != nil {
			return fmt.Errorf("vector_store_config validation failed: %w", err)
		}
	}
	return nil
}
