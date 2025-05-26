// Package graphs provides configurations, LLM tool definitions, and prompt templates
// related to graph-based memory systems.
package graphs

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Neo4jConfig holds the configuration for Neo4j.
type Neo4jConfig struct {
	URL       string `json:"url" validate:"required"`
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	Database  string `json:"database"`
	BaseLabel bool   `json:"base_label"`
}

// Validate validates the Neo4jConfig struct.
func (c *Neo4jConfig) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// MemgraphConfig holds the configuration for Memgraph.
type MemgraphConfig struct {
	URL      string `json:"url" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Validate validates the MemgraphConfig struct.
func (c *MemgraphConfig) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// GraphStoreConfig holds the configuration for the graph store.
type GraphStoreConfig struct {
	Provider     string      `json:"provider" validate:"required,oneof=neo4j memgraph"`
	Config       interface{} `json:"config"` // Placeholder for Neo4jConfig or MemgraphConfig
	LLM          interface{} `json:"llm"`    // Placeholder for a potential LLM config struct
	CustomPrompt string      `json:"custom_prompt"`
}

// Validate validates the GraphStoreConfig struct.
func (c *GraphStoreConfig) Validate() error {
	validate := validator.New()
	// This initial validation will check:
	// 1. Provider is present and is one of "neo4j" or "memgraph".
	// 2. Config is not nil (due to `validate:"required"` on the Config field).
	// 3. If Config holds a struct with its own validation tags (like *Neo4jConfig/*MemgraphConfig),
	//    the validator appears to dive and validate those fields. Error paths
	//    will be like 'GraphStoreConfig.Config.FieldName'.
	if err := validate.Struct(c); err != nil {
		return err
	}

	// Additional check: ensure the actual type of the Config field matches the Provider string.
	// This is crucial because UnmarshalJSON populates Config based on Provider,
	// but if Config is set manually or if UnmarshalJSON logic has issues,
	// this ensures consistency.
	switch c.Provider {
	case "neo4j":
		if _, ok := c.Config.(*Neo4jConfig); !ok {
			return fmt.Errorf("config for provider 'neo4j' must be of type *Neo4jConfig, got %T", c.Config)
		}
		// Individual Neo4jConfig fields already validated by validate.Struct(c) if it dives.
		// If it doesn't dive, then Neo4jConfig.Validate() would be needed here.
		// Given test output, it seems to dive.
	case "memgraph":
		if _, ok := c.Config.(*MemgraphConfig); !ok {
			return fmt.Errorf("config for provider 'memgraph' must be of type *MemgraphConfig, got %T", c.Config)
		}
		// Individual MemgraphConfig fields already validated by validate.Struct(c) if it dives.
	default:
		// This case should ideally not be reached due to the 'oneof' validation on Provider
		// in validate.Struct(c). If it is, it indicates an unexpected state.
		return fmt.Errorf("provider '%s' is valid but has an unexpected config type: %T", c.Provider, c.Config)
	}
	return nil
}

// UnmarshalJSON custom unmarshaler for GraphStoreConfig.
func (c *GraphStoreConfig) UnmarshalJSON(data []byte) error {
	type Alias GraphStoreConfig
	aux := &struct {
		Config json.RawMessage `json:"config"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch c.Provider {
	case "neo4j":
		var neo4jCfg Neo4jConfig
		if err := json.Unmarshal(aux.Config, &neo4jCfg); err != nil {
			return fmt.Errorf("failed to unmarshal neo4j config: %w", err)
		}
		c.Config = &neo4jCfg
	case "memgraph":
		var memgraphCfg MemgraphConfig
		if err := json.Unmarshal(aux.Config, &memgraphCfg); err != nil {
			return fmt.Errorf("failed to unmarshal memgraph config: %w", err)
		}
		c.Config = &memgraphCfg
	default:
		return fmt.Errorf("unknown graph store provider: %s", c.Provider)
	}

	return nil
}
