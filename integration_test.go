package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/divinedev111/hookshot/internal/capture"
	"github.com/divinedev111/hookshot/internal/output"
	"github.com/divinedev111/hookshot/internal/store"
)

func TestFullFlow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	// Set up a forward target
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received":true}`))
	}))
	defer target.Close()

	// Set up the capture server
	var events []*store.Event
	h := capture.NewHandler(s, target.URL, func(e *store.Event) {
		events = append(events, e)
	})
	srv := httptest.NewServer(h)
	defer srv.Close()

	// 1. Send a Stripe webhook
	stripeBody := `{"id":"evt_123","type":"payment_intent.succeeded","data":{}}`
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks?v=1", strings.NewReader(stripeBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "t=1234,v1=abc123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("stripe webhook: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("stripe status = %d", resp.StatusCode)
	}

	// 2. Send a GitHub webhook
	ghBody := `{"ref":"refs/heads/main","commits":[]}`
	req, _ = http.NewRequest("POST", srv.URL+"/hooks", strings.NewReader(ghBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Github-Event", "push")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("github webhook: %v", err)
	}
	resp.Body.Close()

	// 3. Send an unknown webhook
	req, _ = http.NewRequest("POST", srv.URL+"/callback", strings.NewReader(`{"data":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unknown webhook: %v", err)
	}
	resp.Body.Close()

	// Verify callback fired for all 3
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}

	// Verify provider detection
	if events[0].Provider != "stripe" {
		t.Errorf("event 0 provider = %q, want stripe", events[0].Provider)
	}
	if events[0].EventType != "payment_intent.succeeded" {
		t.Errorf("event 0 event_type = %q", events[0].EventType)
	}
	if events[1].Provider != "github" {
		t.Errorf("event 1 provider = %q, want github", events[1].Provider)
	}
	if events[1].EventType != "push" {
		t.Errorf("event 1 event_type = %q", events[1].EventType)
	}
	if events[2].Provider != "" {
		t.Errorf("event 2 provider = %q, want empty", events[2].Provider)
	}

	// Verify forwarding happened
	if events[0].ForwardStatus == nil || *events[0].ForwardStatus != 200 {
		t.Errorf("event 0 forward_status = %v, want 200", events[0].ForwardStatus)
	}

	// --- History: list all events ---
	allEvents, err := s.List(store.ListOpts{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(allEvents) != 3 {
		t.Fatalf("history count = %d, want 3", len(allEvents))
	}

	// --- History: filter by provider ---
	stripeEvents, err := s.List(store.ListOpts{Provider: "stripe"})
	if err != nil {
		t.Fatalf("list stripe: %v", err)
	}
	if len(stripeEvents) != 1 {
		t.Fatalf("stripe count = %d, want 1", len(stripeEvents))
	}

	// --- Inspect ---
	e, err := s.Get(events[0].ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if e.Method != "POST" || e.Path != "/webhooks" {
		t.Errorf("inspect: method=%s path=%s", e.Method, e.Path)
	}
	if e.Query != "v=1" {
		t.Errorf("inspect: query=%s", e.Query)
	}

	detail := output.FormatEventDetail(e, false)
	if !strings.Contains(detail, "Stripe-Signature") {
		t.Error("detail missing Stripe-Signature header")
	}

	detailJSON := output.FormatEventDetail(e, true)
	var m map[string]any
	if err := json.Unmarshal([]byte(detailJSON), &m); err != nil {
		t.Errorf("detail JSON invalid: %v", err)
	}

	// --- Replay ---
	replayTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "payment_intent.succeeded") {
			t.Errorf("replay body missing event type")
		}
		if r.Header.Get("Stripe-Signature") == "" {
			t.Error("replay missing Stripe-Signature header")
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("replayed"))
	}))
	defer replayTarget.Close()

	status, respBody, err := capture.Forward(replayTarget.URL, e.Method, e.Headers, e.Body)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if status != 202 {
		t.Errorf("replay status = %d, want 202", status)
	}
	if string(respBody) != "replayed" {
		t.Errorf("replay body = %q", respBody)
	}

	// --- Export ---
	curl := output.ExportAsCurl(e)
	if !strings.Contains(curl, "curl -X POST") {
		t.Error("curl export missing method")
	}
	if !strings.Contains(curl, "Stripe-Signature") {
		t.Error("curl export missing header")
	}
	if !strings.Contains(curl, "payment_intent.succeeded") {
		t.Error("curl export missing body")
	}

	// --- Format list output ---
	listOut := output.FormatEventList(allEvents, false)
	if !strings.Contains(listOut, "stripe") {
		t.Error("list output missing stripe")
	}
	if !strings.Contains(listOut, "github") {
		t.Error("list output missing github")
	}

	listJSON := output.FormatEventList(allEvents, true)
	var arr []map[string]any
	if err := json.Unmarshal([]byte(listJSON), &arr); err != nil {
		t.Errorf("list JSON invalid: %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("list JSON count = %d, want 3", len(arr))
	}

	// Verify timestamp is recent
	if time.Since(e.ReceivedAt) > 10*time.Second {
		t.Errorf("received_at too old: %v", e.ReceivedAt)
	}

	fmt.Println("Integration test passed: listen -> capture -> history -> inspect -> replay -> export")
}
