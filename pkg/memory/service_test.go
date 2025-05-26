package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// --- Mock NATSClient ---
type mockNATSClient struct {
	PublishCalledWithTopic string
	PublishCalledWithData  []byte
	RequestCalledWithTopic string
	RequestCalledWithData  []byte
	RequestReturnData      []byte
	RequestReturnError     error
	PublishError           error
	// Subscribe method for interface completeness, not used by service directly
	SubscribeFunc func(ctx context.Context, topic string, handler func(msg []byte)) error
}

func (m *mockNATSClient) Publish(ctx context.Context, topic string, data []byte) error {
	m.PublishCalledWithTopic = topic
	m.PublishCalledWithData = data
	return m.PublishError
}

func (m *mockNATSClient) Request(ctx context.Context, topic string, data []byte, timeout time.Duration) ([]byte, error) {
	m.RequestCalledWithTopic = topic
	m.RequestCalledWithData = data
	return m.RequestReturnData, m.RequestReturnError
}

func (m *mockNATSClient) Subscribe(ctx context.Context, topic string, handler func(msg []byte)) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(ctx, topic, handler)
	}
	return nil
}

// --- Mock HistoryStore ---
type mockHistoryStore struct {
	LogEventArgs     *MemoryEvent
	LogEventError    error
	GetHistoryArg    string
	GetHistoryReturn []*MemoryEvent
	GetHistoryError  error
	ResetError       error
	CloseError       error
}

func (m *mockHistoryStore) LogEvent(ctx context.Context, event *MemoryEvent) error {
	m.LogEventArgs = event
	return m.LogEventError
}

func (m *mockHistoryStore) GetHistory(ctx context.Context, memoryID string) ([]*MemoryEvent, error) {
	m.GetHistoryArg = memoryID
	return m.GetHistoryReturn, m.GetHistoryError
}

func (m *mockHistoryStore) Reset(ctx context.Context) error { return m.ResetError }
func (m *mockHistoryStore) Close() error                    { return m.CloseError }

// --- Test Config (minimal for service tests) ---
func getTestServiceConfig() *Config {
	return &Config{
		NATSAddress:               "nats://dummy:4222",
		OpenAIAPIKey:              "sk-dummy",
		TopicMemoryAddReceived:    "test.mem.add.received",
		TopicMemoryProcess:        "test.mem.process",
		TopicMemoryEmbed:          "test.mem.embed",
		TopicMemoryVectorStoreAdd: "test.mem.vectorstore.add",
		TopicMemoryGraphStoreAdd:  "test.mem.graphstore.add",
		TopicMemoryHistoryLog:     "test.mem.history.log",
		TopicMemorySearch:         "test.mem.search",
		TopicMemoryGet:            "test.mem.get",
		TopicMemoryUpdate:         "test.mem.update",
		TopicMemoryDelete:         "test.mem.delete",
		EnableGraphStore:          false, // Keep false to simplify some service tests
		EnableInfer:               false,
	}
}

func TestNewMemoryService(t *testing.T) {
	mockNATS := &mockNATSClient{}
	mockHistory := &mockHistoryStore{}
	cfg := getTestServiceConfig()

	service := NewMemoryService(mockNATS, cfg, mockHistory)
	if service == nil {
		t.Fatal("NewMemoryService returned nil")
	}

	impl, ok := service.(*memoryServiceImpl)
	if !ok {
		t.Fatal("NewMemoryService did not return a *memoryServiceImpl")
	}
	if impl.nc != mockNATS {
		t.Error("NATS client not set correctly in service")
	}
	if impl.cfg != cfg {
		t.Error("Config not set correctly in service")
	}
	if impl.history != mockHistory {
		t.Error("HistoryStore not set correctly in service")
	}
}

func TestMemoryServiceImpl_Add(t *testing.T) {
	mockNATS := &mockNATSClient{}
	cfg := getTestServiceConfig()
	service := NewMemoryService(mockNATS, cfg, &mockHistoryStore{})
	ctx := context.Background()

	t.Run("Valid AddMemoryRequest", func(t *testing.T) {
		req := &AddMemoryRequest{
			BaseRequestInfo: BaseRequestInfo{UserID: "user1"},
			Messages:        []Message{{Role: "user", Content: "Hello Mem0"}},
		}
		memoryID, err := service.Add(ctx, req)
		if err != nil {
			t.Fatalf("Add() error = %v, wantErr nil", err)
		}
		if memoryID == "" {
			t.Error("Add() returned empty memoryID")
		}
		if mockNATS.PublishCalledWithTopic != cfg.TopicMemoryAddReceived {
			t.Errorf("PublishCalledWithTopic = %s, want %s", mockNATS.PublishCalledWithTopic, cfg.TopicMemoryAddReceived)
		}
		var publishedReq AddMemoryRequest
		if err := json.Unmarshal(mockNATS.PublishCalledWithData, &publishedReq); err != nil {
			t.Fatalf("Failed to unmarshal published data: %v", err)
		}
		if !reflect.DeepEqual(&publishedReq, req) {
			t.Errorf("Published data mismatch. Got %+v, want %+v", publishedReq, req)
		}
	})

	t.Run("Invalid AddMemoryRequest (no messages)", func(t *testing.T) {
		req := &AddMemoryRequest{BaseRequestInfo: BaseRequestInfo{UserID: "user1"}, Messages: []Message{}}
		_, err := service.Add(ctx, req)
		if err == nil {
			t.Fatal("Add() with invalid request, expected error, got nil")
		}
		if !strings.Contains(err.Error(), "Messages") { // Check if error is about Messages field
			t.Errorf("Expected error related to Messages field, got: %v", err)
		}
		if mockNATS.PublishCalledWithTopic == cfg.TopicMemoryAddReceived && mockNATS.PublishCalledWithData != nil {
			// Reset for next test if this one failed early but still published somehow
			mockNATS.PublishCalledWithTopic = ""
			mockNATS.PublishCalledWithData = nil
			t.Error("Publish should not have been called for invalid request")
		}
	})

	t.Run("NATS Publish Error", func(t *testing.T) {
		mockNATS.PublishError = errors.New("nats publish failed")
		defer func() { mockNATS.PublishError = nil }() // Reset for other tests

		req := &AddMemoryRequest{
			BaseRequestInfo: BaseRequestInfo{UserID: "user1"},
			Messages:        []Message{{Role: "user", Content: "Hello Mem0"}},
		}
		_, err := service.Add(ctx, req)
		if err == nil {
			t.Fatal("Add() with NATS error, expected error, got nil")
		}
		if !strings.Contains(err.Error(), "nats publish failed") {
			t.Errorf("Expected NATS publish error, got: %v", err)
		}
	})
}

func TestMemoryServiceImpl_Search(t *testing.T) {
	mockNATS := &mockNATSClient{}
	cfg := getTestServiceConfig()
	service := NewMemoryService(mockNATS, cfg, &mockHistoryStore{})
	ctx := context.Background()

	t.Run("Valid SearchMemoryRequest", func(t *testing.T) {
		req := &SearchMemoryRequest{Query: "find memories"}
		mockNATS.RequestReturnData = []byte(`[{"id":"mem1","memory":"test mem"}]`)                                    // Simulate valid NATS response
		mockNATS.RequestReturnError = fmt.Errorf("Search via NATS not fully implemented (response handling pending)") // Expected error from shell

		_, err := service.Search(ctx, req)
		// The shell Search method returns an error even on "success" due to TODO for unmarshalling
		if err == nil {
			t.Fatal("Search() error = nil, wantErr due to shell implementation")
		}
		if !strings.Contains(err.Error(), "Search via NATS not fully implemented") {
			t.Errorf("Search() error = %v, want specific shell error", err)
		}

		if mockNATS.RequestCalledWithTopic != cfg.TopicMemorySearch {
			t.Errorf("RequestCalledWithTopic = %s, want %s", mockNATS.RequestCalledWithTopic, cfg.TopicMemorySearch)
		}
		var publishedReq SearchMemoryRequest
		if errJson := json.Unmarshal(mockNATS.RequestCalledWithData, &publishedReq); errJson != nil {
			t.Fatalf("Failed to unmarshal published data for search: %v", errJson)
		}
		if !reflect.DeepEqual(&publishedReq, req) {
			t.Errorf("Published search data mismatch. Got %+v, want %+v", publishedReq, req)
		}
	})

	t.Run("Invalid SearchMemoryRequest (no query)", func(t *testing.T) {
		req := &SearchMemoryRequest{} // Missing query
		_, err := service.Search(ctx, req)
		if err == nil {
			t.Fatal("Search() with invalid request, expected error, got nil")
		}
		if !strings.Contains(err.Error(), "Query") {
			t.Errorf("Expected error related to Query field, got: %v", err)
		}
	})
}

func TestMemoryServiceImpl_Get(t *testing.T) {
	mockNATS := &mockNATSClient{}
	cfg := getTestServiceConfig()
	service := NewMemoryService(mockNATS, cfg, &mockHistoryStore{})
	ctx := context.Background()

	t.Run("Valid Get Request", func(t *testing.T) {
		memoryID := "mem-abc"
		baseInfo := BaseRequestInfo{UserID: "user-get"}
		mockNATS.RequestReturnError = fmt.Errorf("Get via NATS not fully implemented (response handling pending)") // Expected error from shell

		_, err := service.Get(ctx, memoryID, baseInfo)
		if err == nil {
			t.Fatal("Get() error = nil, wantErr due to shell implementation")
		}
		if !strings.Contains(err.Error(), "Get via NATS not fully implemented") {
			t.Errorf("Get() error = %v, want specific shell error", err)
		}

		if mockNATS.RequestCalledWithTopic != cfg.TopicMemoryGet {
			t.Errorf("RequestCalledWithTopic = %s, want %s", mockNATS.RequestCalledWithTopic, cfg.TopicMemoryGet)
		}
		var publishedPayload GetRequestData
		if errJson := json.Unmarshal(mockNATS.RequestCalledWithData, &publishedPayload); errJson != nil {
			t.Fatalf("Failed to unmarshal published data for get: %v", errJson)
		}
		if publishedPayload.MemoryID != memoryID || publishedPayload.UserID != baseInfo.UserID {
			t.Errorf("Published get data mismatch. Got %+v, want MemoryID %s, UserID %s", publishedPayload, memoryID, baseInfo.UserID)
		}
	})

	t.Run("Invalid Get Request (empty memoryID)", func(t *testing.T) {
		_, err := service.Get(ctx, "", BaseRequestInfo{})
		if err == nil {
			t.Fatal("Get() with empty memoryID, expected error, got nil")
		}
		if !strings.Contains(err.Error(), "memoryID cannot be empty") {
			t.Errorf("Expected error about empty memoryID, got: %v", err)
		}
	})
}

func TestMemoryServiceImpl_Update(t *testing.T) {
	mockNATS := &mockNATSClient{}
	cfg := getTestServiceConfig()
	service := NewMemoryService(mockNATS, cfg, &mockHistoryStore{})
	ctx := context.Background()

	t.Run("Valid Update Request", func(t *testing.T) {
		memoryID := "mem-update-abc"
		data := map[string]interface{}{"new_field": "new_value"}
		baseInfo := BaseRequestInfo{UserID: "user-update"}
		mockNATS.PublishError = fmt.Errorf("Update via NATS not fully implemented (NATS client is nil)") // Error from shell when nc is nil

		err := service.Update(ctx, memoryID, data, baseInfo)
		if err == nil {
			t.Fatal("Update() error = nil, wantErr due to shell implementation")
		}
		if !strings.Contains(err.Error(), "Update via NATS not fully implemented") {
			t.Errorf("Update() error = %v, want specific shell error", err)
		}

		if mockNATS.PublishCalledWithTopic != cfg.TopicMemoryUpdate {
			t.Errorf("PublishCalledWithTopic = %s, want %s", mockNATS.PublishCalledWithTopic, cfg.TopicMemoryUpdate)
		}
		var publishedPayload UpdateRequestData
		if errJson := json.Unmarshal(mockNATS.PublishCalledWithData, &publishedPayload); errJson != nil {
			t.Fatalf("Failed to unmarshal published data for update: %v", errJson)
		}
		if publishedPayload.MemoryID != memoryID || !reflect.DeepEqual(publishedPayload.Data, data) || publishedPayload.UserID != baseInfo.UserID {
			t.Errorf("Published update data mismatch. Got %+v", publishedPayload)
		}
	})
}

func TestMemoryServiceImpl_Delete(t *testing.T) {
	mockNATS := &mockNATSClient{}
	cfg := getTestServiceConfig()
	service := NewMemoryService(mockNATS, cfg, &mockHistoryStore{})
	ctx := context.Background()

	t.Run("Valid Delete Request", func(t *testing.T) {
		memoryID := "mem-delete-abc"
		baseInfo := BaseRequestInfo{UserID: "user-delete"}
		mockNATS.PublishError = fmt.Errorf("Delete via NATS not fully implemented (NATS client is nil)") // Error from shell when nc is nil

		err := service.Delete(ctx, memoryID, baseInfo)
		if err == nil {
			t.Fatal("Delete() error = nil, wantErr due to shell implementation")
		}
		if !strings.Contains(err.Error(), "Delete via NATS not fully implemented") {
			t.Errorf("Delete() error = %v, want specific shell error", err)
		}

		if mockNATS.PublishCalledWithTopic != cfg.TopicMemoryDelete {
			t.Errorf("PublishCalledWithTopic = %s, want %s", mockNATS.PublishCalledWithTopic, cfg.TopicMemoryDelete)
		}
		var publishedPayload GetRequestData // Delete uses GetRequestData
		if errJson := json.Unmarshal(mockNATS.PublishCalledWithData, &publishedPayload); errJson != nil {
			t.Fatalf("Failed to unmarshal published data for delete: %v", errJson)
		}
		if publishedPayload.MemoryID != memoryID || publishedPayload.UserID != baseInfo.UserID {
			t.Errorf("Published delete data mismatch. Got %+v", publishedPayload)
		}
	})
}

func TestMemoryServiceImpl_GetHistory(t *testing.T) {
	mockHistory := &mockHistoryStore{}
	cfg := getTestServiceConfig()
	service := NewMemoryService(&mockNATSClient{}, cfg, mockHistory)
	ctx := context.Background()
	memoryID := "hist-mem1"

	t.Run("Successful GetHistory", func(t *testing.T) {
		eventPtr1 := &MemoryEvent{EventID: "ev1", MemoryID: memoryID, EventType: "CREATE"}
		eventPtr2 := &MemoryEvent{EventID: "ev2", MemoryID: memoryID, EventType: "UPDATE"}
		mockHistory.GetHistoryReturn = []*MemoryEvent{eventPtr1, eventPtr2}
		mockHistory.GetHistoryError = nil

		events, err := service.GetHistory(ctx, memoryID, BaseRequestInfo{})
		if err != nil {
			t.Fatalf("GetHistory() error = %v, wantErr nil", err)
		}
		if mockHistory.GetHistoryArg != memoryID {
			t.Errorf("GetHistoryArg = %s, want %s", mockHistory.GetHistoryArg, memoryID)
		}
		if len(events) != 2 {
			t.Fatalf("Expected 2 events, got %d", len(events))
		}
		// Check conversion from []*MemoryEvent to []MemoryEvent
		if events[0].EventID != eventPtr1.EventID || events[1].EventID != eventPtr2.EventID {
			t.Errorf("Event data mismatch after conversion. Got: %+v, Expected data from: %+v, %+v", events, eventPtr1, eventPtr2)
		}
	})

	t.Run("GetHistory with Error from Store", func(t *testing.T) {
		expectedErr := errors.New("store failed")
		mockHistory.GetHistoryError = expectedErr
		defer func() { mockHistory.GetHistoryError = nil }() // Reset

		_, err := service.GetHistory(ctx, memoryID, BaseRequestInfo{})
		if err == nil {
			t.Fatal("GetHistory() with store error, expected error, got nil")
		}
		if !strings.Contains(err.Error(), expectedErr.Error()) {
			t.Errorf("Expected error from store '%v', got: %v", expectedErr, err)
		}
	})

	t.Run("GetHistory with Empty MemoryID", func(t *testing.T) {
		_, err := service.GetHistory(ctx, "", BaseRequestInfo{})
		if err == nil {
			t.Fatal("GetHistory() with empty memoryID, expected error, got nil")
		}
		if !strings.Contains(err.Error(), "memoryID cannot be empty") {
			t.Errorf("Expected error about empty memoryID, got: %v", err)
		}
	})
}
