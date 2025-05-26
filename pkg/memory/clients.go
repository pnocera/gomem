package memory

import (
	"context"
	"time"

	"github.com/pnocera/gomem/pkg/vectorstores" // Assuming module path
)

// NATSClient defines a minimal interface for NATS publishing and subscribing,
// allowing for easier mocking and integration.
type NATSClient interface {
	Publish(ctx context.Context, topic string, data []byte) error
	Subscribe(ctx context.Context, topic string, handler func(msg []byte)) error // Simplified Subscribe
	Request(ctx context.Context, topic string, data []byte, timeout time.Duration) ([]byte, error)
}

// OpenAIClient placeholder interface defines methods for interacting with an OpenAI-like service.
type OpenAIClient interface {
	ExtractFacts(ctx context.Context, text []string, prompt string) (string, error)
	GetEmbedding(ctx context.Context, text string) ([]float32, error)
	ExtractGraphData(ctx context.Context, text string, prompt string) ([]Entity, []Relation, error)
}

// DgraphClient placeholder interface defines methods for interacting with a Dgraph-like graph database.
// These are simplified for the shell implementation.
type DgraphClient interface {
	Mutate(ctx context.Context, data interface{}) error                              // Simplified
	Query(ctx context.Context, query string, vars map[string]string) ([]byte, error) // Simplified
}

// Ensure vectorstores.VectorStore is available for QdrantWorker.
// This is just to make the import explicit and available if needed directly, though it's
// mainly used as a type for a field in QdrantWorker.
var _ vectorstores.VectorStore = (vectorstores.VectorStore)(nil)
