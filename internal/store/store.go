package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Event represents a captured webhook request.
type Event struct {
	ID              int64
	Method          string
	Path            string
	Query           string
	Headers         map[string][]string
	Body            []byte
	ContentType     string
	Provider        string
	EventType       string
	SourceIP        string
	ReceivedAt      time.Time
	ForwardStatus   *int
	ForwardResponse []byte
}

// ListOpts controls filtering and pagination for listing events.
type ListOpts struct {
	Provider string
	Limit    int
}

// Store provides SQLite-backed persistence for webhook events.
type Store struct {
	db *sql.DB
}

// Open creates or opens a SQLite database and ensures the schema exists.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if _, err := db.Exec(createSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Save inserts an event and sets its ID.
func (s *Store) Save(e *Event) error {
	h, err := json.Marshal(e.Headers)
	if err != nil {
		return fmt.Errorf("marshaling headers: %w", err)
	}

	res, err := s.db.Exec(
		`INSERT INTO events (method, path, query, headers, body, content_type, provider, event_type, source_ip, received_at, forward_status, forward_response)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Method, e.Path, e.Query, h, e.Body, e.ContentType,
		e.Provider, e.EventType, e.SourceIP, e.ReceivedAt,
		e.ForwardStatus, e.ForwardResponse,
	)
	if err != nil {
		return fmt.Errorf("inserting event: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	e.ID = id
	return nil
}

// Get retrieves a single event by ID.
func (s *Store) Get(id int64) (*Event, error) {
	row := s.db.QueryRow(
		`SELECT id, method, path, query, headers, body, content_type, provider, event_type, source_ip, received_at, forward_status, forward_response
		 FROM events WHERE id = ?`, id,
	)
	return scanEvent(row)
}

// List returns events matching the given options.
func (s *Store) List(opts ListOpts) ([]Event, error) {
	q := `SELECT id, method, path, query, headers, body, content_type, provider, event_type, source_ip, received_at, forward_status, forward_response FROM events`
	var args []any

	if opts.Provider != "" {
		q += ` WHERE provider = ?`
		args = append(args, opts.Provider)
	}

	q += ` ORDER BY id DESC`

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	q += ` LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("listing events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		e, err := scanEventRows(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *e)
	}
	return events, rows.Err()
}

// UpdateForward sets the forward status and response for an event.
func (s *Store) UpdateForward(id int64, status int, response []byte) error {
	_, err := s.db.Exec(
		`UPDATE events SET forward_status = ?, forward_response = ? WHERE id = ?`,
		status, response, id,
	)
	if err != nil {
		return fmt.Errorf("updating forward status: %w", err)
	}
	return nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanInto(sc scanner) (*Event, error) {
	var e Event
	var hdrs []byte
	var fwdStatus sql.NullInt64

	err := sc.Scan(
		&e.ID, &e.Method, &e.Path, &e.Query, &hdrs, &e.Body,
		&e.ContentType, &e.Provider, &e.EventType, &e.SourceIP,
		&e.ReceivedAt, &fwdStatus, &e.ForwardResponse,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning event: %w", err)
	}

	if err := json.Unmarshal(hdrs, &e.Headers); err != nil {
		return nil, fmt.Errorf("unmarshaling headers: %w", err)
	}

	if fwdStatus.Valid {
		v := int(fwdStatus.Int64)
		e.ForwardStatus = &v
	}

	return &e, nil
}

func scanEvent(row *sql.Row) (*Event, error) {
	return scanInto(row)
}

func scanEventRows(rows *sql.Rows) (*Event, error) {
	return scanInto(rows)
}
