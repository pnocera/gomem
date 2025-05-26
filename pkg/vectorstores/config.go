package vectorstores

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// QdrantConfig holds configuration specific to Qdrant.
type QdrantConfig struct {
	Address        string `json:"address" validate:"required,url|hostname_port"`
	APIKey         string `json:"api_key,omitempty"`
	CollectionName string `json:"collection_name" validate:"required"`
}

// Validate validates the QdrantConfig struct.
func (c *QdrantConfig) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// VectorStoreConfig holds the configuration for the vector store.
type VectorStoreConfig struct {
	Provider string      `json:"provider" validate:"required,oneof=qdrant"`
	Config   interface{} `json:"config" validate:"required"`
}

// UnmarshalJSON custom unmarshaler for VectorStoreConfig.
func (vsc *VectorStoreConfig) UnmarshalJSON(data []byte) error {
	type VSCProvider struct {
		Provider string          `json:"provider"`
		Config   json.RawMessage `json:"config"`
	}
	var temp VSCProvider
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	vsc.Provider = temp.Provider
	if temp.Config == nil {
		// As per the prompt, error if config field is missing for a specified provider.
		// If provider itself is empty, validator on VectorStoreConfig.Provider will catch it.
		// If provider is specified, config must be present.
		if vsc.Provider != "" { // Check if provider was actually specified in the JSON
			return fmt.Errorf("config field is missing for provider %s", vsc.Provider)
		}
		// If provider is also empty in JSON, let main validator catch missing Config.
		// Setting vsc.Config to nil ensures validate:"required" on it can be checked.
		vsc.Config = nil
		return nil
	}

	switch vsc.Provider {
	case "qdrant":
		var qConfig QdrantConfig
		if err := json.Unmarshal(temp.Config, &qConfig); err != nil {
			return fmt.Errorf("error unmarshalling qdrant config: %w", err)
		}
		vsc.Config = &qConfig
	default:
		// If provider is specified but not "qdrant" (or other supported types in future)
		if vsc.Provider != "" {
			return fmt.Errorf("unsupported vector store provider: %s", vsc.Provider)
		}
		// If provider field was empty/missing in JSON, store raw config.
		// The validator for VectorStoreConfig.Provider will catch the missing provider.
		vsc.Config = temp.Config
	}
	return nil
}

// Validate validates the VectorStoreConfig struct.
func (vsc *VectorStoreConfig) Validate() error {
	validate := validator.New()
	// Validates vsc.Provider ("required", "oneof=qdrant")
	// Validates vsc.Config ("required" - i.e., not nil)
	if err := validate.Struct(vsc); err != nil {
		return err
	}

	// If Config is nil, validate.Struct(vsc) should have caught it.
	// Now, validate the content of the Config based on the provider.
	switch c := vsc.Config.(type) {
	case *QdrantConfig:
		// Provider must be qdrant if Config is QdrantConfig.
		// This is usually ensured by UnmarshalJSON, but good for robustness if Config is set manually.
		if vsc.Provider != "qdrant" {
			return fmt.Errorf("provider is '%s' but config type is *QdrantConfig", vsc.Provider)
		}
		return c.Validate() // Validate the QdrantConfig fields
	default:
		// This case means vsc.Config is not *QdrantConfig.
		// If vsc.Provider is "qdrant", then this is a type mismatch.
		if vsc.Provider == "qdrant" {
			return fmt.Errorf("config for provider '%s' is of unexpected type %T", vsc.Provider, vsc.Config)
		}
		// If vsc.Provider is not "qdrant", it should have been caught by the 'oneof' tag
		// in validate.Struct(vsc). If it somehow wasn't (e.g. provider is empty string),
		// this indicates an unknown config type for an unspecified or unsupported provider.
		return fmt.Errorf("unknown config type (%T) for provider '%s'", vsc.Config, vsc.Provider)
	}
}
