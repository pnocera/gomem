package graphs

import (
	"reflect"
	"testing"
)

func TestAddMemoryToolGraphDefinition(t *testing.T) {
	tool := AddMemoryToolGraph

	if tool.Type != "function" {
		t.Errorf("Expected Type to be 'function', got '%s'", tool.Type)
	}

	if tool.Function.Name != "add_graph_memory" {
		t.Errorf("Expected Function.Name to be 'add_graph_memory', got '%s'", tool.Function.Name)
	}

	expectedDescription := "Add a new graph memory to the knowledge graph. This function creates a new relationship between two nodes, potentially creating new nodes if they don't exist."
	if tool.Function.Description != expectedDescription {
		t.Errorf("Expected Function.Description to be '%s', got '%s'", expectedDescription, tool.Function.Description)
	}

	if tool.Function.Parameters.Type != "object" {
		t.Errorf("Expected Function.Parameters.Type to be 'object', got '%s'", tool.Function.Parameters.Type)
	}

	expectedProperties := map[string]ToolParameterProperty{
		"source":           {Type: "string", Description: "The identifier of the source node. This could be an entity, concept, or any piece of information that forms the starting point of the relationship."},
		"destination":      {Type: "string", Description: "The identifier of the destination node. This is the entity, concept, or piece of information that the source node is related to."},
		"relationship":     {Type: "string", Description: "The type of relationship between the source and destination nodes. This word/phrase describes how the source and destination are connected."},
		"source_type":      {Type: "string", Description: "The type or category of the source node. For example, 'Person', 'Location', 'Concept'. Helps in organizing and querying the graph."},
		"destination_type": {Type: "string", Description: "The type or category of the destination node. For example, 'Event', 'Organization', 'Skill'. Helps in organizing and querying the graph."},
	}

	if !reflect.DeepEqual(tool.Function.Parameters.Properties, expectedProperties) {
		t.Errorf("Expected Function.Parameters.Properties to be '%v', got '%v'", expectedProperties, tool.Function.Parameters.Properties)
	}

	if property, ok := tool.Function.Parameters.Properties["source"]; !ok {
		t.Error("Expected 'source' property to exist")
	} else {
		if property.Type != "string" {
			t.Errorf("Expected 'source' property Type to be 'string', got '%s'", property.Type)
		}
		if property.Description != expectedProperties["source"].Description {
			t.Errorf("Expected 'source' property Description to be '%s', got '%s'", expectedProperties["source"].Description, property.Description)
		}
	}

	expectedRequired := []string{"source", "destination", "relationship", "source_type", "destination_type"}
	if !reflect.DeepEqual(tool.Function.Parameters.Required, expectedRequired) {
		t.Errorf("Expected Function.Parameters.Required to be '%v', got '%v'", expectedRequired, tool.Function.Parameters.Required)
	}
}

func TestExtractEntitiesToolDefinition(t *testing.T) {
	tool := ExtractEntitiesTool

	if tool.Type != "function" {
		t.Errorf("Expected Type to be 'function', got '%s'", tool.Type)
	}

	if tool.Function.Name != "extract_entities" {
		t.Errorf("Expected Function.Name to be 'extract_entities', got '%s'", tool.Function.Name)
	}

	if _, ok := tool.Function.Parameters.Properties["entities"]; !ok {
		t.Error("Expected 'entities' property to exist")
	}

	entitiesProperty := tool.Function.Parameters.Properties["entities"]
	if entitiesProperty.Type != "array" {
		t.Errorf("Expected 'entities' property Type to be 'array', got '%s'", entitiesProperty.Type)
	}

	expectedDescription := "A list of entities extracted from the text. Each entity should be a sub-list or array containing two strings: [entity_name, entity_type]. For example, [['John Doe', 'Person'], ['New York', 'Location']]."
	if entitiesProperty.Description != expectedDescription {
		t.Errorf("Expected 'entities' property Description to be '%s', got '%s'", expectedDescription, entitiesProperty.Description)
	}

	expectedRequired := []string{"entities"}
	if !reflect.DeepEqual(tool.Function.Parameters.Required, expectedRequired) {
		t.Errorf("Expected Function.Parameters.Required to be '%v', got '%v'", expectedRequired, tool.Function.Parameters.Required)
	}
}

func TestNoopToolDefinition(t *testing.T) {
	tool := NoopTool

	if tool.Type != "function" {
		t.Errorf("Expected Type to be 'function', got '%s'", tool.Type)
	}

	if tool.Function.Name != "noop" {
		t.Errorf("Expected Function.Name to be 'noop', got '%s'", tool.Function.Name)
	}

	if len(tool.Function.Parameters.Properties) != 0 {
		t.Errorf("Expected Function.Parameters.Properties to be empty, got '%v'", tool.Function.Parameters.Properties)
	}

	if len(tool.Function.Parameters.Required) != 0 {
		t.Errorf("Expected Function.Parameters.Required to be empty, got '%v'", tool.Function.Parameters.Required)
	}
}
