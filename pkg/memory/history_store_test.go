package memory

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"
)

// Helper to create a new in-memory SQLite store for each test
func newTestSQLiteHistoryStore(t *testing.T) (*SQLiteHistoryStore, string) {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "test_history_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file for test DB: %v", err)
	}
	dbPath := tmpfile.Name()
	if err := tmpfile.Close(); err != nil {
		os.Remove(dbPath) // Clean up if close fails
		t.Fatalf("Failed to close temp file: %v", err)
	}

	store, err := NewSQLiteHistoryStore(dbPath)
	if err != nil {
		os.Remove(dbPath) // Clean up if NewSQLiteHistoryStore fails
		t.Fatalf("NewSQLiteHistoryStore() error = %v", err)
	}
	return store, dbPath
}

func TestNewSQLiteHistoryStore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		store, dbPath := newTestSQLiteHistoryStore(t)
		defer os.Remove(dbPath) // Ensure cleanup
		defer store.Close()

		if store.db == nil {
			t.Error("Expected db to be initialized, got nil")
		}
		if store.dbPath != dbPath {
			t.Errorf("Expected dbPath to be '%s', got '%s'", dbPath, store.dbPath)
		}

		var tableName string
		err := store.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='history';").Scan(&tableName)
		if err != nil {
			t.Errorf("Failed to query for history table: %v", err)
		}
		if tableName != "history" {
			t.Errorf("Expected table 'history' to exist, found '%s'", tableName)
		}
	})

	t.Run("Failure - Invalid DSN (e.g., unwriteable path)", func(t *testing.T) {
		// This test is OS-dependent and hard to make reliably fail for all environments.
		// Creating a file in a non-existent directory.
		invalidPath := "/some_nonexistent_directory_for_testing_github.com/pnocera/gomemtest.db"
		_, err := NewSQLiteHistoryStore(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid DSN (unwriteable path), got nil")
			// Attempt to clean up if a file was somehow created, though unlikely.
			os.Remove(invalidPath)
		} else {
			// Check for a common error substring, though this can vary.
			// e.g., "unable to open database file", "no such file or directory"
			// For now, just ensuring an error occurred is sufficient for this shell test.
			t.Logf("Received expected error for invalid DSN: %v", err)
		}
	})
}

func TestLogEventAndGetHistory(t *testing.T) {
	store, dbPath := newTestSQLiteHistoryStore(t)
	defer os.Remove(dbPath)
	defer store.Close()

	ctx := context.Background()
	memID1 := "memory-1"
	memID2 := "memory-2"

	// Using Truncate to avoid SQLite's varying timestamp precision issues vs Go's time.Time
	ts1 := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Millisecond)
	ts2 := time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Millisecond)
	ts3 := time.Now().UTC().Truncate(time.Millisecond)

	event1 := &MemoryEvent{
		MemoryID:  memID1,
		EventType: "CREATE",
		Timestamp: ts1,
		UserID:    "user-a",
		Details:   map[string]interface{}{"key1": "value1", "num": 123.45, "nested": map[string]interface{}{"nk": "nv"}},
	}
	event2 := &MemoryEvent{
		MemoryID:  memID1,
		EventType: "UPDATE",
		Timestamp: ts2,
		UserID:    "user-a",
		NewMemory: "updated content",
		Details:   map[string]interface{}{"key2": true},
	}
	event3 := &MemoryEvent{ // Event with no details
		MemoryID:    memID2,
		EventType:   "SEARCH",
		Timestamp:   ts3,
		UserID:      "user-b",
		SearchQuery: "find this",
	}
	event4 := &MemoryEvent{ // Event with nil details
		MemoryID:  memID2,
		EventType: "DELETE",
		Timestamp: ts3.Add(time.Second), // Ensure different timestamp
		UserID:    "user-b",
		Details:   nil,
	}

	// Log events
	if err := store.LogEvent(ctx, event1); err != nil {
		t.Fatalf("LogEvent() error for event1 = %v", err)
	}
	if event1.EventID == "" { // EventID should be populated by LogEvent
		t.Fatal("event1.EventID was not populated")
	}
	originalEvent1ID := event1.EventID // Save for later comparison

	if err := store.LogEvent(ctx, event2); err != nil {
		t.Fatalf("LogEvent() error for event2 = %v", err)
	}
	if err := store.LogEvent(ctx, event3); err != nil {
		t.Fatalf("LogEvent() error for event3 = %v", err)
	}
	if err := store.LogEvent(ctx, event4); err != nil {
		t.Fatalf("LogEvent() error for event4 = %v", err)
	}

	t.Run("GetHistory for memoryID1", func(t *testing.T) {
		history, err := store.GetHistory(ctx, memID1)
		if err != nil {
			t.Fatalf("GetHistory() error = %v", err)
		}
		if len(history) != 2 {
			t.Fatalf("Expected 2 events for memoryID1, got %d events: %+v", len(history), history)
		}

		// The query orders by timestamp, so history[0] should be event1, history[1] should be event2
		retrievedEvent1 := history[0]
		if retrievedEvent1.EventID != originalEvent1ID { // Compare with the ID populated by LogEvent
			t.Errorf("Event 1 EventID mismatch. Got %s, expected %s", retrievedEvent1.EventID, originalEvent1ID)
		}
		if !reflect.DeepEqual(retrievedEvent1.Details, event1.Details) {
			t.Errorf("Event 1 Details mismatch. Got %+v, expected %+v", retrievedEvent1.Details, event1.Details)
		}
		// Check a nested value in details
		if nestedMap, ok := retrievedEvent1.Details["nested"].(map[string]interface{}); !ok || nestedMap["nk"] != "nv" {
			t.Errorf("Event 1 nested detail mismatch. Got %+v", retrievedEvent1.Details["nested"])
		}

		retrievedEvent2 := history[1]
		if !reflect.DeepEqual(retrievedEvent2.Details, event2.Details) {
			t.Errorf("Event 2 Details mismatch. Got %+v, expected %+v", retrievedEvent2.Details, event2.Details)
		}
	})

	t.Run("GetHistory for memoryID2", func(t *testing.T) {
		history, err := store.GetHistory(ctx, memID2)
		if err != nil {
			t.Fatalf("GetHistory() error = %v", err)
		}
		if len(history) != 2 { // Now expecting event3 and event4
			t.Fatalf("Expected 2 events for memoryID2, got %d", len(history))
		}
		// event3 should have empty (not nil) details map after retrieval due to current GetHistory logic
		if history[0].EventType != event3.EventType {
			t.Errorf("Expected first event for memID2 to be of type '%s', got '%s'", event3.EventType, history[0].EventType)
		}
		if len(history[0].Details) != 0 { // Details was nil, GetHistory makes it an empty map
			t.Errorf("Event 3 (no details) Details mismatch. Got %+v, expected empty map", history[0].Details)
		}
		// event4 should also have empty (not nil) details map
		if history[1].EventType != event4.EventType {
			t.Errorf("Expected second event for memID2 to be of type '%s', got '%s'", event4.EventType, history[1].EventType)
		}
		if len(history[1].Details) != 0 { // Details was nil, GetHistory makes it an empty map
			t.Errorf("Event 4 (nil details) Details mismatch. Got %+v, expected empty map", history[1].Details)
		}
	})

	t.Run("GetHistory for non-existent memoryID", func(t *testing.T) {
		history, err := store.GetHistory(ctx, "non-existent-id")
		if err != nil {
			t.Fatalf("GetHistory() error = %v", err)
		}
		if len(history) != 0 {
			t.Errorf("Expected 0 events for non-existent memoryID, got %d", len(history))
		}
	})
}

func TestResetHistoryStore(t *testing.T) {
	store, dbPath := newTestSQLiteHistoryStore(t)
	defer os.Remove(dbPath)
	defer store.Close()
	ctx := context.Background()

	event := &MemoryEvent{MemoryID: "reset-test", EventType: "TEST", Timestamp: time.Now()}
	if err := store.LogEvent(ctx, event); err != nil {
		t.Fatalf("LogEvent() error = %v", err)
	}

	history, err := store.GetHistory(ctx, "reset-test")
	if err != nil || len(history) != 1 {
		t.Fatalf("Expected 1 event before Reset, got %d, err: %v", len(history), err)
	}

	if err := store.Reset(ctx); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	historyAfterReset, err := store.GetHistory(ctx, "reset-test")
	if err != nil {
		t.Fatalf("GetHistory() after Reset error = %v", err)
	}
	if len(historyAfterReset) != 0 {
		t.Errorf("Expected 0 events after Reset, got %d", len(historyAfterReset))
	}

	// Verify table still exists and is usable
	newEvent := &MemoryEvent{MemoryID: "reset-test-2", EventType: "POST-RESET", Timestamp: time.Now()}
	if err := store.LogEvent(ctx, newEvent); err != nil {
		t.Fatalf("LogEvent() after Reset error = %v", err)
	}
	historyAfterNewLog, err := store.GetHistory(ctx, "reset-test-2")
	if err != nil || len(historyAfterNewLog) != 1 {
		t.Fatalf("Expected 1 event after logging post-Reset, got %d, err: %v", len(historyAfterNewLog), err)
	}
	if historyAfterNewLog[0].EventType != "POST-RESET" {
		t.Errorf("Unexpected event type after re-logging: %s", historyAfterNewLog[0].EventType)
	}
}

func TestCloseHistoryStore(t *testing.T) {
	store, dbPath := newTestSQLiteHistoryStore(t)
	originalPath := dbPath // Keep original path for os.Remove verification

	// First Close
	if err := store.Close(); err != nil {
		t.Errorf("First Close() returned error: %v", err)
	}
	// After the custom Close in newTestSQLiteHistoryStore, the file should be removed.
	// However, the deferred os.Remove(dbPath) in newTestSQLiteHistoryStore will also run.
	// The test here is about the store's Close method.
	// The helper's Close method is what removes the file.
	if store.db != nil {
		t.Error("Expected store.db to be nil after Close()")
	}
	if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
		t.Errorf("Expected temp DB file %s to be removed by the helper's Close, but it still exists or error checking: %v", originalPath, err)
	}

	// Multiple Closes on the already closed store (via helper's Close)
	if err := store.Close(); err != nil {
		t.Errorf("Second Close() returned error: %v", err)
	}

	// Test operation after close
	errAfterClose := store.LogEvent(context.Background(), &MemoryEvent{EventType: "AFTER_CLOSE", EventID: "event-after-close"})
	expectedErrorMsg := "SQLiteHistoryStore is closed"
	if errAfterClose == nil {
		t.Errorf("Expected error '%s' from LogEvent after Close, got nil", expectedErrorMsg)
	} else if errAfterClose.Error() != expectedErrorMsg {
		t.Errorf("Expected error '%s' from LogEvent after Close, got: %v", expectedErrorMsg, errAfterClose)
	}
}
