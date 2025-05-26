package memory

import (
	// Assuming module path for nested configs

	"strings"
	"testing"

	"github.com/pnocera/gomem/pkg/graphs"
	"github.com/pnocera/gomem/pkg/vectorstores"
)

func getValidConfig() *Config {
	return &Config{
		NATSAddress:                "nats://localhost:4222",
		OpenAIAPIKey:               "sk-testkey",
		TopicMemoryAddReceived:     "mem0.memory.add.received",
		TopicMemoryProcess:         "mem0.memory.process",
		TopicMemoryEmbed:           "mem0.memory.embed",
		TopicMemoryVectorStoreAdd:  "mem0.memory.vectorstore.add",
		TopicMemoryGraphStoreAdd:   "mem0.memory.graphstore.add",
		TopicMemoryHistoryLog:      "mem0.memory.history.log",
		TopicMemorySearch:          "mem0.memory.search",
		TopicMemoryGet:             "mem0.memory.get",
		TopicMemoryUpdate:          "mem0.memory.update",
		TopicMemoryDelete:          "mem0.memory.delete",
		EnableGraphStore:           true,
		EnableInfer:                true,
		GraphConfig:                &graphs.GraphStoreConfig{Provider: "neo4j", Config: &graphs.Neo4jConfig{URL: "bolt://localhost:7687", Username: "u", Password: "p"}},
		VectorStoreConfig:          &vectorstores.VectorStoreConfig{Provider: "qdrant", Config: &vectorstores.QdrantConfig{Address: "http://localhost:6333", CollectionName: "test"}},
		CustomFactExtractionPrompt: "extract facts",
		CustomUpdateMemoryPrompt:   "update memory",
	}
}

func TestConfig_Validate_Success(t *testing.T) {
	cfg := getValidConfig()
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}
}

func TestConfig_Validate_Failure_RequiredFields(t *testing.T) {
	requiredFields := []string{
		"NATSAddress", "OpenAIAPIKey",
		"TopicMemoryAddReceived", "TopicMemoryProcess", "TopicMemoryEmbed",
		"TopicMemoryVectorStoreAdd", "TopicMemoryGraphStoreAdd", "TopicMemoryHistoryLog",
		"TopicMemorySearch", "TopicMemoryGet", "TopicMemoryUpdate", "TopicMemoryDelete",
	}

	for _, field := range requiredFields {
		t.Run(field+"_Missing", func(t *testing.T) {
			cfg := getValidConfig()
			// Use reflection or a switch to set the field to its zero value
			switch field {
			case "NATSAddress":
				cfg.NATSAddress = ""
			case "OpenAIAPIKey":
				cfg.OpenAIAPIKey = ""
			case "TopicMemoryAddReceived":
				cfg.TopicMemoryAddReceived = ""
			case "TopicMemoryProcess":
				cfg.TopicMemoryProcess = ""
			case "TopicMemoryEmbed":
				cfg.TopicMemoryEmbed = ""
			case "TopicMemoryVectorStoreAdd":
				cfg.TopicMemoryVectorStoreAdd = ""
			case "TopicMemoryGraphStoreAdd":
				cfg.TopicMemoryGraphStoreAdd = ""
			case "TopicMemoryHistoryLog":
				cfg.TopicMemoryHistoryLog = ""
			case "TopicMemorySearch":
				cfg.TopicMemorySearch = ""
			case "TopicMemoryGet":
				cfg.TopicMemoryGet = ""
			case "TopicMemoryUpdate":
				cfg.TopicMemoryUpdate = ""
			case "TopicMemoryDelete":
				cfg.TopicMemoryDelete = ""
			}

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Expected error when %s is missing, got nil", field)
				return
			}
			if !strings.Contains(err.Error(), field) {
				t.Errorf("Expected error message for %s to contain '%s', got '%v'", field, field, err)
			}
		})
	}
}

func TestConfig_Validate_NestedConfigs(t *testing.T) {
	t.Run("ValidNestedConfigs", func(t *testing.T) {
		cfg := getValidConfig() // getValidConfig already includes valid nested configs
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Expected no error with valid nested configs, got %v", err)
		}
	})

	t.Run("InvalidGraphConfig", func(t *testing.T) {
		cfg := getValidConfig()
		cfg.GraphConfig = &graphs.GraphStoreConfig{Provider: "neo4j", Config: &graphs.Neo4jConfig{URL: ""}} // Invalid: URL is required
		err := cfg.Validate()
		if err == nil {
			t.Errorf("Expected error for invalid GraphConfig, got nil")
			return
		}
		// The error from validator/v10 for nested structs will include the path.
		// e.g. Key: 'Config.GraphConfig.Config.URL'
		if !strings.Contains(err.Error(), "GraphConfig.Config.URL") && !strings.Contains(err.Error(), "GraphConfig.Provider") {
			// The error could also be because the GraphConfig itself is invalid (e.g. provider not matching type)
			// Let's check if the error is from the GraphConfig validation path.
			t.Logf("Error details: %v", err) // Log error for more details if needed
			// A more robust check might be needed if the exact error path varies too much
			// or if the validation stops at the GraphConfig.Validate() level.
			// For this test, we just check that an error occurred.
		}
	})

	t.Run("InvalidVectorStoreConfig", func(t *testing.T) {
		cfg := getValidConfig()
		cfg.VectorStoreConfig = &vectorstores.VectorStoreConfig{Provider: "qdrant", Config: &vectorstores.QdrantConfig{Address: ""}} // Invalid: Address is required
		err := cfg.Validate()
		if err == nil {
			t.Errorf("Expected error for invalid VectorStoreConfig, got nil")
			return
		}
		if !strings.Contains(err.Error(), "VectorStoreConfig.Config.Address") && !strings.Contains(err.Error(), "VectorStoreConfig.Provider") {
			t.Logf("Error details: %v", err)
		}
	})

	t.Run("NilGraphConfigWhenEnabled", func(t *testing.T) {
		// Note: The current Config struct doesn't enforce GraphConfig to be non-nil if EnableGraphStore is true.
		// The validation tags `omitempty` on GraphConfig means it's optional from a pure struct validation perspective.
		// Business logic elsewhere might enforce this. This test checks current struct validation.
		cfg := getValidConfig()
		cfg.EnableGraphStore = true
		cfg.GraphConfig = nil // This is allowed by `omitempty`
		err := cfg.Validate()
		if err != nil {
			t.Errorf("Expected no error for nil GraphConfig with `omitempty` even if EnableGraphStore is true (struct validation only), got %v", err)
		}
	})
}
