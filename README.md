# gomem

A Go implementation of a memory system inspired by mem0.

## Overview

gomem is a memory management system for AI applications, providing vector and graph-based storage for conversational memory. It enables AI agents to store, retrieve, and reason about past interactions and knowledge.

## Features

- **Dual Storage System**: Combines vector embeddings and knowledge graphs
- **Memory Management**: Add, search, update, and delete memories
- **Event History**: Track operations performed on memories
- **NATS Integration**: Asynchronous processing via NATS messaging
- **Extensible Architecture**: Support for different vector and graph database providers

## Components

### Memory Service

The core service that handles memory operations:
- Add new memories from conversations
- Search for relevant memories
- Track history of memory operations
- Update and delete existing memories

### Vector Stores

Storage for vector embeddings with semantic search capabilities:
- Qdrant implementation
- Common interface for adding other providers

### Graph Stores

Knowledge graph storage for structured relationships:
- Neo4j implementation
- Memgraph support
- Relationship extraction and graph updates

## Getting Started

### Prerequisites

- Go 1.23.1 or higher
- NATS server (for production use)
- Vector database (Qdrant recommended)
- Graph database (Neo4j or Memgraph)

### Installation

```bash
go get github.com/pnocera/gomem
```

### Basic Usage

```go
import (
    "context"
    "github.com/pnocera/gomem/pkg/memory"
)

// Initialize memory service
historyStore, _ := memory.NewSQLiteHistoryStore("memory.db")
memoryService := memory.NewMemoryService(natsClient, &memConfig, historyStore)

// Add a memory
memoryID, err := memoryService.Add(context.Background(), &memory.AddMemoryRequest{
    BaseRequestInfo: memory.BaseRequestInfo{UserID: "user-123"},
    Messages: []memory.Message{
        {Role: "user", Content: "What's the capital of France?"},
        {Role: "assistant", Content: "The capital of France is Paris."},
    },
})

// Search memories
results, _ := memoryService.Search(context.Background(), &memory.SearchMemoryRequest{
    BaseRequestInfo: memory.BaseRequestInfo{UserID: "user-123"},
    Query: "France capital",
    Limit: 5,
})
```

## Configuration

The system is configured through the `memory.Config` struct:

```go
config := memory.Config{
    NATSAddress: "nats://localhost:4222",
    OpenAIAPIKey: "your-api-key",
    EnableGraphStore: true,
    EnableInfer: true,
    GraphConfig: &graphs.GraphStoreConfig{
        Provider: "neo4j",
        Config: &graphs.Neo4jConfig{
            URL: "bolt://localhost:7687",
            Username: "neo4j",
            Password: "password",
        },
    },
    VectorStoreConfig: &vectorstores.VectorStoreConfig{
        Provider: "qdrant",
        Config: &vectorstores.QdrantConfig{
            Address: "http://localhost:6333",
            CollectionName: "memories",
        },
    },
}
```

## Example

See `cmd/example/main.go` for a complete example of using the memory service.

## License

[License information]
