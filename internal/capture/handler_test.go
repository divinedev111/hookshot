package capture

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/divinedev111/hookshot/internal/store"
)

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestHandlerCapturesRequest(t *testing.T) {
	s := openTestStore(t)
	var captured *store.Event

	h := NewHandler(s, "", func(e *store.Event) {
		captured = e
	})

	body := `{"type":"payment_intent.succeeded"}`
	req := httptest.NewRequest("POST", "/webhooks?foo=bar", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "t=123,v1=abc")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	if captured == nil {
		t.Fatal("callback not called")
	}
	if captured.Method != "POST" {
		t.Errorf("method = %q", captured.Method)
	}
	if captured.Path != "/webhooks" {
		t.Errorf("path = %q", captured.Path)
	}
	if captured.Query != "foo=bar" {
		t.Errorf("query = %q", captured.Query)
	}
	if captured.Provider != "stripe" {
		t.Errorf("provider = %q", captured.Provider)
	}
	if captured.EventType != "payment_intent.succeeded" {
		t.Errorf("event_type = %q", captured.EventType)
	}
	if captured.ID == 0 {
		t.Error("expected event to have ID after save")
	}

	got, err := s.Get(captured.ID)
	if err != nil {
		t.Fatalf("get from store: %v", err)
	}
	if got.Provider != "stripe" {
		t.Errorf("stored provider = %q", got.Provider)
	}
}

func TestHandlerWithForward(t *testing.T) {
	s := openTestStore(t)

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("accepted"))
	}))
	defer target.Close()

	var captured *store.Event
	h := NewHandler(s, target.URL, func(e *store.Event) {
		captured = e
	})

	req := httptest.NewRequest("POST", "/hook", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if captured == nil {
		t.Fatal("callback not called")
	}
	if captured.ForwardStatus == nil || *captured.ForwardStatus != 202 {
		t.Errorf("forward_status = %v, want 202", captured.ForwardStatus)
	}

	got, err := s.Get(captured.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ForwardStatus == nil || *got.ForwardStatus != 202 {
		t.Errorf("stored forward_status = %v, want 202", got.ForwardStatus)
	}
}
