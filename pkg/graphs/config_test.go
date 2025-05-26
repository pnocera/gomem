package graphs

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestGraphStoreConfig_UnmarshalJSON(t *testing.T) {
	t.Run("Successful unmarshal with Neo4j provider", func(t *testing.T) {
		jsonData := []byte(`{
			"provider": "neo4j",
			"config": {
				"url": "bolt://localhost:7687",
				"username": "neo4j_user",
				"password": "neo4j_password",
				"database": "neo4j",
				"base_label": true
			},
			"custom_prompt": "test prompt"
		}`)
		var cfg GraphStoreConfig
		err := json.Unmarshal(jsonData, &cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.Provider != "neo4j" {
			t.Errorf("Expected provider to be 'neo4j', got '%s'", cfg.Provider)
		}
		if cfg.CustomPrompt != "test prompt" {
			t.Errorf("Expected custom_prompt to be 'test prompt', got '%s'", cfg.CustomPrompt)
		}

		neo4jCfg, ok := cfg.Config.(*Neo4jConfig)
		if !ok {
			t.Fatalf("Expected Config to be of type *Neo4jConfig, got %T", cfg.Config)
		}
		if neo4jCfg.URL != "bolt://localhost:7687" {
			t.Errorf("Expected Neo4j URL to be 'bolt://localhost:7687', got '%s'", neo4jCfg.URL)
		}
		if neo4jCfg.Username != "neo4j_user" {
			t.Errorf("Expected Neo4j Username to be 'neo4j_user', got '%s'", neo4jCfg.Username)
		}
		if neo4jCfg.Password != "neo4j_password" {
			t.Errorf("Expected Neo4j Password to be 'neo4j_password', got '%s'", neo4jCfg.Password)
		}
		if neo4jCfg.Database != "neo4j" {
			t.Errorf("Expected Neo4j Database to be 'neo4j', got '%s'", neo4jCfg.Database)
		}
		if !neo4jCfg.BaseLabel {
			t.Errorf("Expected Neo4j BaseLabel to be true, got %v", neo4jCfg.BaseLabel)
		}
	})

	t.Run("Successful unmarshal with Memgraph provider", func(t *testing.T) {
		jsonData := []byte(`{
			"provider": "memgraph",
			"config": {
				"url": "bolt://localhost:7688",
				"username": "memgraph_user",
				"password": "memgraph_password"
			}
		}`)
		var cfg GraphStoreConfig
		err := json.Unmarshal(jsonData, &cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.Provider != "memgraph" {
			t.Errorf("Expected provider to be 'memgraph', got '%s'", cfg.Provider)
		}

		memgraphCfg, ok := cfg.Config.(*MemgraphConfig)
		if !ok {
			t.Fatalf("Expected Config to be of type *MemgraphConfig, got %T", cfg.Config)
		}
		if memgraphCfg.URL != "bolt://localhost:7688" {
			t.Errorf("Expected Memgraph URL to be 'bolt://localhost:7688', got '%s'", memgraphCfg.URL)
		}
		if memgraphCfg.Username != "memgraph_user" {
			t.Errorf("Expected Memgraph Username to be 'memgraph_user', got '%s'", memgraphCfg.Username)
		}
		if memgraphCfg.Password != "memgraph_password" {
			t.Errorf("Expected Memgraph Password to be 'memgraph_password', got '%s'", memgraphCfg.Password)
		}
	})

	t.Run("Unmarshal failure with unsupported provider", func(t *testing.T) {
		jsonData := []byte(`{
			"provider": "unsupported_provider",
			"config": {}
		}`)
		var cfg GraphStoreConfig
		err := json.Unmarshal(jsonData, &cfg)
		if err == nil {
			t.Fatal("Expected an error for unsupported provider, got nil")
		}
		expectedErrorMsg := "unknown graph store provider: unsupported_provider"
		if err.Error() != expectedErrorMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
		}
	})

	t.Run("Unmarshal failure if config is not a map for neo4j", func(t *testing.T) {
		jsonData := []byte(`{
			"provider": "neo4j",
			"config": "not_a_map"
		}`)
		var cfg GraphStoreConfig
		err := json.Unmarshal(jsonData, &cfg)
		if err == nil {
			t.Fatal("Expected an error for invalid config type, got nil")
		}
		expectedErrorMsg := "failed to unmarshal neo4j config: json: cannot unmarshal string into Go value of type graphs.Neo4jConfig"
		if err.Error() != expectedErrorMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
		}
	})

	t.Run("Unmarshal failure if config is not a map for memgraph", func(t *testing.T) {
		jsonData := []byte(`{
			"provider": "memgraph",
			"config": "not_a_map"
		}`)
		var cfg GraphStoreConfig
		err := json.Unmarshal(jsonData, &cfg)
		if err == nil {
			t.Fatal("Expected an error for invalid config type, got nil")
		}
		expectedErrorMsg := "failed to unmarshal memgraph config: json: cannot unmarshal string into Go value of type graphs.MemgraphConfig"
		if err.Error() != expectedErrorMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
		}
	})
}

func TestNeo4jConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Neo4jConfig
		wantErr bool
	}{
		{"Valid config", Neo4jConfig{URL: "url", Username: "user", Password: "pass"}, false},
		{"Missing URL", Neo4jConfig{Username: "user", Password: "pass"}, true},
		{"Missing Username", Neo4jConfig{URL: "url", Password: "pass"}, true},
		{"Missing Password", Neo4jConfig{URL: "url", Username: "user"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Neo4jConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemgraphConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  MemgraphConfig
		wantErr bool
	}{
		{"Valid config", MemgraphConfig{URL: "url", Username: "user", Password: "pass"}, false},
		{"Missing URL", MemgraphConfig{Username: "user", Password: "pass"}, true},
		{"Missing Username", MemgraphConfig{URL: "url", Password: "pass"}, true},
		{"Missing Password", MemgraphConfig{URL: "url", Username: "user"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MemgraphConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGraphStoreConfig_Validate(t *testing.T) {
	validNeo4j := &Neo4jConfig{URL: "url", Username: "user", Password: "password"}
	invalidNeo4j := &Neo4jConfig{Username: "user", Password: "password"} // Missing URL

	validMemgraph := &MemgraphConfig{URL: "url", Username: "user", Password: "password"}
	invalidMemgraph := &MemgraphConfig{Username: "user", Password: "password"} // Missing URL

	tests := []struct {
		name        string
		config      GraphStoreConfig
		wantErr     bool
		errContains string
	}{
		{"Valid Neo4j config", GraphStoreConfig{Provider: "neo4j", Config: validNeo4j}, false, ""},
		{"Valid Memgraph config", GraphStoreConfig{Provider: "memgraph", Config: validMemgraph}, false, ""},
		{"Missing Provider", GraphStoreConfig{Config: validNeo4j}, true, "Key: 'GraphStoreConfig.Provider' Error:Field validation for 'Provider' failed on the 'required' tag"},
		{"Unsupported Provider", GraphStoreConfig{Provider: "invalid_provider", Config: validNeo4j}, true, "Key: 'GraphStoreConfig.Provider' Error:Field validation for 'Provider' failed on the 'oneof' tag"},
		{"Invalid Neo4j config (field validation)", GraphStoreConfig{Provider: "neo4j", Config: invalidNeo4j}, true, "Key: 'GraphStoreConfig.Config.URL' Error:Field validation for 'URL' failed on the 'required' tag"},
		{"Invalid Memgraph config (field validation)", GraphStoreConfig{Provider: "memgraph", Config: invalidMemgraph}, true, "Key: 'GraphStoreConfig.Config.URL' Error:Field validation for 'URL' failed on the 'required' tag"},
		{"Config is nil", GraphStoreConfig{Provider: "neo4j", Config: nil}, true, "config for provider 'neo4j' must be of type *Neo4jConfig, got <nil>"},
		{"Config is wrong type for provider (neo4j)", GraphStoreConfig{Provider: "neo4j", Config: validMemgraph}, true, "config for provider 'neo4j' must be of type *Neo4jConfig, got *graphs.MemgraphConfig"},
		{"Config is wrong type for provider (memgraph)", GraphStoreConfig{Provider: "memgraph", Config: validNeo4j}, true, "config for provider 'memgraph' must be of type *MemgraphConfig, got *graphs.Neo4jConfig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("GraphStoreConfig.Validate() error = '%v', wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GraphStoreConfig.Validate() error = '%v', wantErrContains '%v'", err, tt.errContains)
				}
			}
		})
	}
}

// Helper to allow direct string contains check for specific validation errors
func (s *GraphStoreConfig) ValidateAndCheckError(expectedErrorContent string) error {
	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		if expectedErrorContent != "" && !strings.Contains(err.Error(), expectedErrorContent) {
			return fmt.Errorf("expected error to contain '%s', but got: %w", expectedErrorContent, err)
		}
		return err
	}

	switch cfg := s.Config.(type) {
	case *Neo4jConfig:
		err = cfg.Validate()
		if err != nil {
			return fmt.Errorf("neo4j config validation failed: %w", err)
		}
	case *MemgraphConfig:
		err = cfg.Validate()
		if err != nil {
			return fmt.Errorf("memgraph config validation failed: %w", err)
		}
	default:
		return fmt.Errorf("unknown graph store config type: %T", s.Config)
	}
	return nil
}
