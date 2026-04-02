package store

import (
	"path/filepath"
	"testing"
	"time"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSaveAndGet(t *testing.T) {
	s := openTestStore(t)

	e := &Event{
		Method:      "POST",
		Path:        "/webhooks",
		Query:       "foo=bar",
		Headers:     map[string][]string{"Content-Type": {"application/json"}},
		Body:        []byte(`{"type":"test"}`),
		ContentType: "application/json",
		Provider:    "stripe",
		EventType:   "payment_intent.succeeded",
		SourceIP:    "127.0.0.1",
		ReceivedAt:  time.Now().UTC().Truncate(time.Second),
	}

	if err := s.Save(e); err != nil {
		t.Fatalf("save: %v", err)
	}
	if e.ID == 0 {
		t.Fatal("expected non-zero ID after save")
	}

	got, err := s.Get(e.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.Method != e.Method {
		t.Errorf("method = %q, want %q", got.Method, e.Method)
	}
	if got.Path != e.Path {
		t.Errorf("path = %q, want %q", got.Path, e.Path)
	}
	if got.Provider != e.Provider {
		t.Errorf("provider = %q, want %q", got.Provider, e.Provider)
	}
	if got.EventType != e.EventType {
		t.Errorf("event_type = %q, want %q", got.EventType, e.EventType)
	}
	if got.ForwardStatus != nil {
		t.Errorf("forward_status = %v, want nil", got.ForwardStatus)
	}
}

func TestUpdateForward(t *testing.T) {
	s := openTestStore(t)

	e := &Event{
		Method:     "POST",
		Path:       "/hook",
		Headers:    map[string][]string{},
		ReceivedAt: time.Now().UTC(),
	}
	if err := s.Save(e); err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := s.UpdateForward(e.ID, 200, []byte("ok")); err != nil {
		t.Fatalf("update forward: %v", err)
	}

	got, err := s.Get(e.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ForwardStatus == nil || *got.ForwardStatus != 200 {
		t.Errorf("forward_status = %v, want 200", got.ForwardStatus)
	}
	if string(got.ForwardResponse) != "ok" {
		t.Errorf("forward_response = %q, want %q", got.ForwardResponse, "ok")
	}
}

func TestListWithFilters(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC()

	events := []Event{
		{Method: "POST", Path: "/a", Headers: map[string][]string{}, Provider: "stripe", ReceivedAt: now},
		{Method: "POST", Path: "/b", Headers: map[string][]string{}, Provider: "github", ReceivedAt: now},
		{Method: "POST", Path: "/c", Headers: map[string][]string{}, Provider: "stripe", ReceivedAt: now},
	}
	for i := range events {
		if err := s.Save(&events[i]); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	tests := []struct {
		name string
		opts ListOpts
		want int
	}{
		{"all", ListOpts{}, 3},
		{"filter stripe", ListOpts{Provider: "stripe"}, 2},
		{"filter github", ListOpts{Provider: "github"}, 1},
		{"limit 1", ListOpts{Limit: 1}, 1},
		{"filter nonexistent", ListOpts{Provider: "slack"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.List(tt.opts)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(got) != tt.want {
				t.Errorf("got %d events, want %d", len(got), tt.want)
			}
		})
	}
}

func TestListOrderDescending(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC()

	for i := 0; i < 3; i++ {
		e := &Event{Method: "POST", Path: "/x", Headers: map[string][]string{}, ReceivedAt: now}
		if err := s.Save(e); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	got, err := s.List(ListOpts{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	for i := 1; i < len(got); i++ {
		if got[i].ID >= got[i-1].ID {
			t.Errorf("events not in descending order: id %d >= %d", got[i].ID, got[i-1].ID)
		}
	}
}

func TestGetNotFound(t *testing.T) {
	s := openTestStore(t)
	_, err := s.Get(999)
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
}
