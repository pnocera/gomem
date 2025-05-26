package graphs

// ToolParameterProperty defines the structure for individual parameters within a tool.
type ToolParameterProperty struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ToolParameters defines the overall parameter structure for a tool.
type ToolParameters struct {
	Type       string                           `json:"type"`
	Properties map[string]ToolParameterProperty `json:"properties"`
	Required   []string                         `json:"required,omitempty"`
}

// FunctionDefinition defines the function that a tool can execute.
type FunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

// Tool defines the structure for an LLM tool.
type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// AddMemoryToolGraph defines the tool for adding a new graph memory.
var AddMemoryToolGraph = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "add_graph_memory",
		Description: "Add a new graph memory to the knowledge graph. This function creates a new relationship between two nodes, potentially creating new nodes if they don't exist.",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"source":           {Type: "string", Description: "The identifier of the source node. This could be an entity, concept, or any piece of information that forms the starting point of the relationship."},
				"destination":      {Type: "string", Description: "The identifier of the destination node. This is the entity, concept, or piece of information that the source node is related to."},
				"relationship":     {Type: "string", Description: "The type of relationship between the source and destination nodes. This word/phrase describes how the source and destination are connected."},
				"source_type":      {Type: "string", Description: "The type or category of the source node. For example, 'Person', 'Location', 'Concept'. Helps in organizing and querying the graph."},
				"destination_type": {Type: "string", Description: "The type or category of the destination node. For example, 'Event', 'Organization', 'Skill'. Helps in organizing and querying the graph."},
			},
			Required: []string{"source", "destination", "relationship", "source_type", "destination_type"},
		},
	},
}

// UpdateMemoryToolGraph defines the tool for updating an existing graph memory.
var UpdateMemoryToolGraph = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "update_graph_memory",
		Description: "Update an existing graph memory in the knowledge graph. This function updates the relationship or properties of an existing node or relationship.",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"source":           {Type: "string", Description: "The identifier of the source node of the memory to be updated."},
				"destination":      {Type: "string", Description: "The identifier of the destination node of the memory to be updated."},
				"relationship":     {Type: "string", Description: "The new relationship type to update to."},
				"source_type":      {Type: "string", Description: "The new type of the source node."},
				"destination_type": {Type: "string", Description: "The new type of the destination node."},
			},
			Required: []string{"source", "destination", "relationship", "source_type", "destination_type"},
		},
	},
}

// NoopTool defines a tool that does nothing.
var NoopTool = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "noop",
		Description: "No operation. Use this function when no specific action is needed based on the input, or when the information is not relevant to any other tool.",
		Parameters: ToolParameters{
			Type:       "object",
			Properties: map[string]ToolParameterProperty{},
		},
	},
}

// RelationsTool defines the tool for extracting relationships from text.
var RelationsTool = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "extract_relations",
		Description: "Extracts relationships from a given text and represents them as a graph. Each relationship is a triplet: (source_node, destination_node, relationship_type).",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"relations": {
					Type:        "array",
					Description: "A list of relationships extracted from the text. Each relationship should be a sub-list or array containing three strings: [source_node, destination_node, relationship_type]. For example, [['person_A', 'person_B', 'knows'], ['person_A', 'company_X', 'works_at']].",
				},
			},
			Required: []string{"relations"},
		},
	},
}

// ExtractEntitiesTool defines the tool for extracting entities from text.
var ExtractEntitiesTool = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "extract_entities",
		Description: "Extracts entities from a given text. Each entity has a name and a type.",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"entities": {
					Type:        "array",
					Description: "A list of entities extracted from the text. Each entity should be a sub-list or array containing two strings: [entity_name, entity_type]. For example, [['John Doe', 'Person'], ['New York', 'Location']].",
				},
			},
			Required: []string{"entities"},
		},
	},
}

// UpdateMemoryStructToolGraph defines the tool for updating memories from a structured input.
var UpdateMemoryStructToolGraph = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "update_memory_struct",
		Description: "Update existing memories in the graph based on a structured input. This function takes a list of memory structures, each specifying a source, destination, and relationship, and updates them in the graph.",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"memories": {
					Type:        "array",
					Description: "A list of memory structures to update. Each memory should be an object with 'source', 'destination', and 'relationship' keys. For example, [{'source': 'entity_A', 'destination': 'entity_B', 'relationship': 'updated_relation'}].",
				},
			},
			Required: []string{"memories"},
		},
	},
}

// AddMemoryStructToolGraph defines the tool for adding new memories from a structured input.
var AddMemoryStructToolGraph = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "add_memory_struct",
		Description: "Add new memories to the graph from a structured input. This function takes a list of memory structures, each specifying a source, destination, and relationship, and adds them to the graph.",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"memories": {
					Type:        "array",
					Description: "A list of memory structures to add. Each memory should be an object with 'source', 'destination', and 'relationship' keys. For example, [{'source': 'entity_A', 'destination': 'entity_B', 'relationship': 'new_relation'}].",
				},
			},
			Required: []string{"memories"},
		},
	},
}

// NoopStructTool defines a tool that does nothing for structured data.
var NoopStructTool = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "noop_struct",
		Description: "No operation for structured data. Use this function when no specific action is needed based on the structured input.",
		Parameters: ToolParameters{
			Type:       "object",
			Properties: map[string]ToolParameterProperty{},
		},
	},
}

// RelationsStructTool defines the tool for extracting relationships from structured data.
var RelationsStructTool = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "extract_relations_struct",
		Description: "Extracts relationships from a given structured data and represents them as a graph. Each relationship is a triplet: (source_node, destination_node, relationship_type).",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"relations": {
					Type:        "array",
					Description: "A list of relationships extracted from the structured data. Each relationship should be an object with 'source', 'destination', and 'relationship' keys. For example, [{'source': 'entity_A', 'destination': 'entity_B', 'relationship': 'relation_type'}].",
				},
			},
			Required: []string{"relations"},
		},
	},
}

// ExtractEntitiesStructTool defines the tool for extracting entities from structured data.
var ExtractEntitiesStructTool = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "extract_entities_struct",
		Description: "Extracts entities from a given structured data. Each entity has a name and a type.",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"entities": {
					Type:        "array",
					Description: "A list of entities extracted from the structured data. Each entity should be an object with 'name' and 'type' keys. For example, [{'name': 'John Doe', 'type': 'Person'}, {'name': 'New York', 'type': 'Location'}].",
				},
			},
			Required: []string{"entities"},
		},
	},
}

// DeleteMemoryStructToolGraph defines the tool for deleting memories from a structured input.
var DeleteMemoryStructToolGraph = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "delete_memory_struct",
		Description: "Delete memories from the graph based on a structured input. This function takes a list of memory structures, each specifying a source, destination, and relationship to delete.",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"memories": {
					Type:        "array",
					Description: "A list of memory structures to delete. Each memory should be an object with 'source', 'destination', and 'relationship' keys. For example, [{'source': 'entity_A', 'destination': 'entity_B', 'relationship': 'relation_to_delete'}].",
				},
			},
			Required: []string{"memories"},
		},
	},
}

// DeleteMemoryToolGraph defines the tool for deleting a specific graph memory.
var DeleteMemoryToolGraph = Tool{
	Type: "function",
	Function: FunctionDefinition{
		Name:        "delete_graph_memory",
		Description: "Delete a specific memory from the knowledge graph. This function removes a relationship between two nodes, and potentially the nodes themselves if they become orphaned.",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolParameterProperty{
				"source":       {Type: "string", Description: "The identifier of the source node of the memory to be deleted."},
				"destination":  {Type: "string", Description: "The identifier of the destination node of the memory to be deleted."},
				"relationship": {Type: "string", Description: "The type of relationship to be deleted."},
				// "source_type":  {Type: "string", Description: "The type of the source node. Optional, but helps in disambiguation."},
				// "destination_type": {Type: "string", Description: "The type of the destination node. Optional, but helps in disambiguation."},
			},
			Required: []string{"source", "destination", "relationship"},
		},
	},
}
