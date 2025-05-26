package memory

import (
	"context"
	"testing"
	"time"

	"github.com/pnocera/gomem/pkg/vectorstores" // Assuming module path for vectorstores.VectorStore
)

// --- Mock VectorStore for QdrantWorker tests ---
type mockVectorStore struct {
	CreateCollectionFunc    func(name string, vectorSize int, distanceMetric string) error
	DeleteCollectionFunc    func(name string) error
	ListCollectionsFunc     func() ([]string, error)
	CollectionInfoFunc      func(name string) (*vectorstores.CollectionInfo, error)
	ResetCollectionFunc     func(name string, vectorSize int, distanceMetric string) error
	InsertVectorsFunc       func(collectionName string, vectors []vectorstores.VectorInput) error
	UpdateVectorPayloadFunc func(collectionName string, vectorID string, payload map[string]interface{}) error
	GetVectorFunc           func(collectionName string, vectorID string) (*vectorstores.SearchResult, error)
	DeleteVectorsFunc       func(collectionName string, vectorIDs []string) error
	SearchFunc              func(collectionName string, queryEmbedding []float32, limit int, filter *vectorstores.QueryFilter) ([]vectorstores.SearchResult, error)
	ListVectorsFunc         func(collectionName string, limit int, offset uint64, filter *vectorstores.QueryFilter) ([]vectorstores.SearchResult, error)
}

func (m *mockVectorStore) CreateCollection(name string, vectorSize int, distanceMetric string) error {
	if m.CreateCollectionFunc != nil {
		return m.CreateCollectionFunc(name, vectorSize, distanceMetric)
	}
	return nil
}
func (m *mockVectorStore) DeleteCollection(name string) error {
	if m.DeleteCollectionFunc != nil {
		return m.DeleteCollectionFunc(name)
	}
	return nil
}
func (m *mockVectorStore) ListCollections() ([]string, error) {
	if m.ListCollectionsFunc != nil {
		return m.ListCollectionsFunc()
	}
	return nil, nil
}
func (m *mockVectorStore) CollectionInfo(name string) (*vectorstores.CollectionInfo, error) {
	if m.CollectionInfoFunc != nil {
		return m.CollectionInfoFunc(name)
	}
	return nil, nil
}
func (m *mockVectorStore) ResetCollection(name string, vectorSize int, distanceMetric string) error {
	if m.ResetCollectionFunc != nil {
		return m.ResetCollectionFunc(name, vectorSize, distanceMetric)
	}
	return nil
}
func (m *mockVectorStore) InsertVectors(collectionName string, vectors []vectorstores.VectorInput) error {
	if m.InsertVectorsFunc != nil {
		return m.InsertVectorsFunc(collectionName, vectors)
	}
	return nil
}
func (m *mockVectorStore) UpdateVectorPayload(collectionName string, vectorID string, payload map[string]interface{}) error {
	if m.UpdateVectorPayloadFunc != nil {
		return m.UpdateVectorPayloadFunc(collectionName, vectorID, payload)
	}
	return nil
}
func (m *mockVectorStore) GetVector(collectionName string, vectorID string) (*vectorstores.SearchResult, error) {
	if m.GetVectorFunc != nil {
		return m.GetVectorFunc(collectionName, vectorID)
	}
	return nil, nil
}
func (m *mockVectorStore) DeleteVectors(collectionName string, vectorIDs []string) error {
	if m.DeleteVectorsFunc != nil {
		return m.DeleteVectorsFunc(collectionName, vectorIDs)
	}
	return nil
}
func (m *mockVectorStore) Search(collectionName string, queryEmbedding []float32, limit int, filter *vectorstores.QueryFilter) ([]vectorstores.SearchResult, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(collectionName, queryEmbedding, limit, filter)
	}
	return nil, nil
}
func (m *mockVectorStore) ListVectors(collectionName string, limit int, offset uint64, filter *vectorstores.QueryFilter) ([]vectorstores.SearchResult, error) {
	if m.ListVectorsFunc != nil {
		return m.ListVectorsFunc(collectionName, limit, offset, filter)
	}
	return nil, nil
}

// TestNewQdrantWorker ensures worker can be created.
func TestNewQdrantWorker(t *testing.T) {
	cfg := &Config{TopicMemoryVectorStoreAdd: "test.topic.qdrant"} // Minimal config
	mockNATS := &mockNATSClient{}
	mockVS := &mockVectorStore{}
	worker := NewQdrantWorker(mockNATS, cfg, mockVS)
	if worker == nil {
		t.Errorf("NewQdrantWorker returned nil")
	}
	if worker.nc != mockNATS {
		t.Error("QdrantWorker: NATS client not set correctly")
	}
	if worker.cfg != cfg {
		t.Error("QdrantWorker: Config not set correctly")
	}
	if worker.vs != mockVS {
		t.Error("QdrantWorker: VectorStore client not set correctly")
	}
}

// TestQdrantWorker_StartStop ensures Start can be called and respects context cancellation.
func TestQdrantWorker_StartStop(t *testing.T) {
	cfg := &Config{TopicMemoryVectorStoreAdd: "test.qdrant.startstop"}
	mockNATS := &mockNATSClient{}
	mockVS := &mockVectorStore{}
	worker := NewQdrantWorker(mockNATS, cfg, mockVS)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond) // Increased timeout
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Start(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Worker Start returned unexpected error: %v, expected nil on context done", err)
		}
	case <-time.After(400 * time.Millisecond): // Increased test safety timeout
		t.Errorf("Worker Start did not return after context cancellation")
	}
}
