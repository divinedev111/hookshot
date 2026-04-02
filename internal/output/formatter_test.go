package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/divinedev111/hookshot/internal/store"
)

func makeEvent() *store.Event {
	status := 200
	return &store.Event{
		ID:            1,
		Method:        "POST",
		Path:          "/webhooks",
		Query:         "foo=bar",
		Headers:       map[string][]string{"Content-Type": {"application/json"}, "Stripe-Signature": {"t=123"}},
		Body:          []byte(`{"type":"payment_intent.succeeded"}`),
		ContentType:   "application/json",
		Provider:      "stripe",
		EventType:     "payment_intent.succeeded",
		SourceIP:      "127.0.0.1",
		ReceivedAt:    time.Now().UTC(),
		ForwardStatus: &status,
	}
}

func TestFormatEventContainsKey(t *testing.T) {
	e := makeEvent()
	out := FormatEvent(e, false)

	for _, want := range []string{"#1", "stripe", "POST", "/webhooks"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %s", want, out)
		}
	}
}

func TestFormatEventJSON(t *testing.T) {
	e := makeEvent()
	out := FormatEvent(e, true)

	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["provider"] != "stripe" {
		t.Errorf("provider = %v", m["provider"])
	}
}

func TestFormatEventDetailContainsHeaders(t *testing.T) {
	e := makeEvent()
	out := FormatEventDetail(e, false)

	if !strings.Contains(out, "Stripe-Signature") {
		t.Errorf("output missing header: %s", out)
	}
	if !strings.Contains(out, "payment_intent.succeeded") {
		t.Errorf("output missing body content")
	}
}

func TestFormatEventDetailJSON(t *testing.T) {
	e := makeEvent()
	out := FormatEventDetail(e, true)

	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["method"] != "POST" {
		t.Errorf("method = %v", m["method"])
	}
	if _, ok := m["headers"]; !ok {
		t.Error("missing headers in detail JSON")
	}
}

func TestFormatEventListEmpty(t *testing.T) {
	out := FormatEventList(nil, false)
	if !strings.Contains(out, "No events") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestFormatEventListJSON(t *testing.T) {
	events := []store.Event{*makeEvent()}
	out := FormatEventList(events, true)

	var arr []map[string]any
	if err := json.Unmarshal([]byte(out), &arr); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(arr) != 1 {
		t.Errorf("expected 1 event, got %d", len(arr))
	}
}

func TestFormatEventUnknownProvider(t *testing.T) {
	e := &store.Event{
		ID:         2,
		Method:     "POST",
		Path:       "/callback",
		Headers:    map[string][]string{},
		ReceivedAt: time.Now().UTC(),
	}
	out := FormatEvent(e, false)
	if !strings.Contains(out, "[unknown]") {
		t.Errorf("expected [unknown] provider: %s", out)
	}
}
