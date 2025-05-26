package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// HistoryStore defines the interface for logging and retrieving memory events.
type HistoryStore interface {
	// LogEvent records a memory event.
	LogEvent(ctx context.Context, event *MemoryEvent) error

	// GetHistory retrieves all events for a specific memory ID, ordered by timestamp.
	GetHistory(ctx context.Context, memoryID string) ([]*MemoryEvent, error)

	// Reset clears all history.
	Reset(ctx context.Context) error

	// Close closes any underlying database connections.
	Close() error
}

// SQLiteHistoryStore implements the HistoryStore interface using SQLite.
type SQLiteHistoryStore struct {
	db     *sql.DB
	dbPath string
	mu     sync.RWMutex // For protecting schema changes or multi-step operations
}

// Compile-time check to ensure *SQLiteHistoryStore satisfies the HistoryStore interface.
var _ HistoryStore = (*SQLiteHistoryStore)(nil)

// NewSQLiteHistoryStore creates a new SQLiteHistoryStore instance.
func NewSQLiteHistoryStore(dataSourceName string) (*SQLiteHistoryStore, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	store := &SQLiteHistoryStore{
		db:     db,
		dbPath: dataSourceName,
	}

	if err := store._createHistoryTable(); err != nil {
		store.Close()
		return nil, fmt.Errorf("failed to create history table: %w", err)
	}

	return store, nil
}

// _createHistoryTable creates the history table if it doesn't already exist.
func (s *SQLiteHistoryStore) _createHistoryTable() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS history (
		event_id TEXT PRIMARY KEY,
		memory_id TEXT,
		event_type TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		user_id TEXT,
		agent_id TEXT,
		run_id TEXT,
		actor_id TEXT,
		old_memory TEXT,
		new_memory TEXT,
		search_query TEXT,
		details TEXT
	);`

	createMemoryIDIndexSQL := `CREATE INDEX IF NOT EXISTS idx_history_memory_id ON history (memory_id);`
	createTimestampIndexSQL := `CREATE INDEX IF NOT EXISTS idx_history_timestamp ON history (timestamp);`

	_, err := s.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to execute create history table statement: %w", err)
	}

	_, err = s.db.Exec(createMemoryIDIndexSQL)
	if err != nil {
		return fmt.Errorf("failed to create memory_id index: %w", err)
	}

	_, err = s.db.Exec(createTimestampIndexSQL)
	if err != nil {
		return fmt.Errorf("failed to create timestamp index: %w", err)
	}

	return nil
}

// LogEvent records a memory event.
func (s *SQLiteHistoryStore) LogEvent(ctx context.Context, event *MemoryEvent) error {
	s.mu.Lock() // Ensure exclusive access for preparing statement and inserting
	defer s.mu.Unlock()

	if s.db == nil {
		return fmt.Errorf("SQLiteHistoryStore is closed")
	}

	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	detailsJSON, err := json.Marshal(event.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal event details to JSON: %w", err)
	}

	stmt, err := s.db.PrepareContext(ctx, `
		INSERT INTO history (
			event_id, memory_id, event_type, timestamp, user_id, agent_id, 
			run_id, actor_id, old_memory, new_memory, search_query, details
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement for history: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		event.EventID,
		sql.NullString{String: event.MemoryID, Valid: event.MemoryID != ""},
		event.EventType,
		event.Timestamp,
		sql.NullString{String: event.UserID, Valid: event.UserID != ""},
		sql.NullString{String: event.AgentID, Valid: event.AgentID != ""},
		sql.NullString{String: event.RunID, Valid: event.RunID != ""},
		sql.NullString{String: event.ActorID, Valid: event.ActorID != ""},
		sql.NullString{String: event.OldMemory, Valid: event.OldMemory != ""},
		sql.NullString{String: event.NewMemory, Valid: event.NewMemory != ""},
		sql.NullString{String: event.SearchQuery, Valid: event.SearchQuery != ""},
		string(detailsJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to execute insert statement for history: %w", err)
	}

	return nil
}

// GetHistory retrieves all events for a specific memory ID, ordered by timestamp.
func (s *SQLiteHistoryStore) GetHistory(ctx context.Context, memoryID string) ([]*MemoryEvent, error) {
	s.mu.RLock() // Use RLock for read operations
	defer s.mu.RUnlock()

	query := `
		SELECT event_id, memory_id, event_type, timestamp, user_id, agent_id,
		       run_id, actor_id, old_memory, new_memory, search_query, details
		FROM history
		WHERE memory_id = ?
		ORDER BY timestamp ASC
	`
	rows, err := s.db.QueryContext(ctx, query, memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query history for memory_id %s: %w", memoryID, err)
	}
	defer rows.Close()

	var events []*MemoryEvent
	for rows.Next() {
		event := &MemoryEvent{}
		var detailsJSON sql.NullString // Use sql.NullString for potentially NULL details
		var memID, userID, agentID, runID, actorID, oldMem, newMem, searchQuery sql.NullString

		err := rows.Scan(
			&event.EventID,
			&memID, // Scan into sql.NullString
			&event.EventType,
			&event.Timestamp,
			&userID,
			&agentID,
			&runID,
			&actorID,
			&oldMem,
			&newMem,
			&searchQuery,
			&detailsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history row: %w", err)
		}

		// Assign from sql.NullString to string fields
		event.MemoryID = memID.String
		event.UserID = userID.String
		event.AgentID = agentID.String
		event.RunID = runID.String
		event.ActorID = actorID.String
		event.OldMemory = oldMem.String
		event.NewMemory = newMem.String
		event.SearchQuery = searchQuery.String

		if detailsJSON.Valid && detailsJSON.String != "" {
			if err := json.Unmarshal([]byte(detailsJSON.String), &event.Details); err != nil {
				// Log or handle error if details are crucial, but don't fail the whole query
				// For example, you could set event.Details to a map indicating an unmarshal error.
				// For now, we'll let it be nil if unmarshalling fails.
				// Consider logging this error: fmt.Printf("Warning: failed to unmarshal details for event %s: %v\n", event.EventID, err)
				event.Details = make(map[string]interface{}) // Ensure Details is not nil
				event.Details["error"] = "failed to unmarshal details: " + err.Error()
			}
		} else {
			event.Details = make(map[string]interface{}) // Ensure Details is not nil if details were NULL or empty
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history rows: %w", err)
	}

	return events, nil
}

// Reset clears all history.
func (s *SQLiteHistoryStore) Reset(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dropTableSQL := `DROP TABLE IF EXISTS history;`
	_, err := s.db.ExecContext(ctx, dropTableSQL)
	if err != nil {
		return fmt.Errorf("failed to drop history table: %w", err)
	}

	// Re-create the table
	// Note: _createHistoryTable already acquires its own lock, but since Reset holds the lock,
	// it's okay. If _createHistoryTable was called without Reset holding the lock,
	// it would manage its own concurrency.
	return s._createHistoryTable()
}

// Close closes any underlying database connections.
func (s *SQLiteHistoryStore) Close() error {
	s.mu.Lock() // Ensure exclusive access for closing
	defer s.mu.Unlock()

	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			return fmt.Errorf("failed to close sqlite database: %w", err)
		}
		s.db = nil // Mark as closed
		return nil
	}
	return nil // Already closed or not initialized
}
