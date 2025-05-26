package vectorstores

import "fmt" // For "not implemented" errors

// QdrantStore implements the VectorStore interface for Qdrant.
// This is a shell implementation.
type QdrantStore struct {
	// client interface{} // Placeholder for actual Qdrant client
	// defaultCollection string
}

// Compile-time check to ensure *QdrantStore satisfies the VectorStore interface.
var _ VectorStore = (*QdrantStore)(nil)

// NewQdrantStore creates a new QdrantStore instance (shell function).
// func NewQdrantStore(config *QdrantConfig) (VectorStore, error) {
//     return nil, fmt.Errorf("NewQdrantStore not implemented")
// }

func (s *QdrantStore) CreateCollection(name string, vectorSize int, distanceMetric string) error {
	return fmt.Errorf("CreateCollection not implemented")
}

func (s *QdrantStore) DeleteCollection(name string) error {
	return fmt.Errorf("DeleteCollection not implemented")
}

func (s *QdrantStore) ListCollections() ([]string, error) {
	return nil, fmt.Errorf("ListCollections not implemented")
}

func (s *QdrantStore) CollectionInfo(name string) (*CollectionInfo, error) {
	return nil, fmt.Errorf("CollectionInfo not implemented")
}

func (s *QdrantStore) ResetCollection(name string, vectorSize int, distanceMetric string) error {
	return fmt.Errorf("ResetCollection not implemented")
}

func (s *QdrantStore) InsertVectors(collectionName string, vectors []VectorInput) error {
	return fmt.Errorf("InsertVectors not implemented")
}

func (s *QdrantStore) UpdateVectorPayload(collectionName string, vectorID string, payload map[string]interface{}) error {
	return fmt.Errorf("UpdateVectorPayload not implemented")
}

func (s *QdrantStore) GetVector(collectionName string, vectorID string) (*SearchResult, error) {
	return nil, fmt.Errorf("GetVector not implemented")
}

func (s *QdrantStore) DeleteVectors(collectionName string, vectorIDs []string) error {
	return fmt.Errorf("DeleteVectors not implemented")
}

func (s *QdrantStore) Search(collectionName string, queryEmbedding []float32, limit int, filter *QueryFilter) ([]SearchResult, error) {
	return nil, fmt.Errorf("Search not implemented")
}

func (s *QdrantStore) ListVectors(collectionName string, limit int, offset uint64, filter *QueryFilter) ([]SearchResult, error) {
	return nil, fmt.Errorf("ListVectors not implemented")
}
