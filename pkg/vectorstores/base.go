// Package vectorstores defines common interfaces, data structures, and configurations
// for interacting with various vector database providers.
package vectorstores

// VectorInput represents a single data point to be inserted into the vector store.
type VectorInput struct {
	ID        string                 `json:"id"`
	Embedding []float32              `json:"embedding"`
	Payload   map[string]interface{} `json:"payload"`
}

// SearchResult represents a single search result from the vector store.
type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

// QueryFilter defines filters to be applied during a search operation.
type QueryFilter struct {
	UserID   string                 `json:"user_id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CollectionInfo holds information about a vector store collection.
type CollectionInfo struct {
	Name       string `json:"name"`
	VectorSize int    `json:"vector_size"`
	PointCount uint64 `json:"point_count"`
}

// VectorStore defines the common interface for interacting with a vector database.
type VectorStore interface {
	CreateCollection(name string, vectorSize int, distanceMetric string) error
	DeleteCollection(name string) error
	ListCollections() ([]string, error)
	CollectionInfo(name string) (*CollectionInfo, error)
	ResetCollection(name string, vectorSize int, distanceMetric string) error

	InsertVectors(collectionName string, vectors []VectorInput) error
	UpdateVectorPayload(collectionName string, vectorID string, payload map[string]interface{}) error
	GetVector(collectionName string, vectorID string) (*SearchResult, error)
	DeleteVectors(collectionName string, vectorIDs []string) error
	Search(collectionName string, queryEmbedding []float32, limit int, filter *QueryFilter) ([]SearchResult, error)
	ListVectors(collectionName string, limit int, offset uint64, filter *QueryFilter) ([]SearchResult, error)
}
