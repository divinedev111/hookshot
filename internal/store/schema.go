package store

const createSchema = `
CREATE TABLE IF NOT EXISTS events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	method TEXT NOT NULL,
	path TEXT NOT NULL,
	query TEXT,
	headers TEXT NOT NULL,
	body BLOB,
	content_type TEXT,
	provider TEXT,
	event_type TEXT,
	source_ip TEXT,
	received_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	forward_status INTEGER,
	forward_response BLOB
);
`
