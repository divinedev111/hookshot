package output

import (
	"strings"
	"testing"

	"github.com/divinedev111/hookshot/internal/store"
)

func TestExportAsCurl(t *testing.T) {
	e := &store.Event{
		Method:  "POST",
		Path:    "/webhooks",
		Query:   "v=1",
		Headers: map[string][]string{"Content-Type": {"application/json"}, "Stripe-Signature": {"t=123,v1=abc"}},
		Body:    []byte(`{"id":"evt_123","type":"payment_intent.succeeded"}`),
	}

	out := ExportAsCurl(e)

	checks := []string{
		"curl -X POST",
		"'http://localhost/webhooks?v=1'",
		"-H 'Content-Type: application/json'",
		"-H 'Stripe-Signature: t=123,v1=abc'",
		`-d '{"id":"evt_123","type":"payment_intent.succeeded"}'`,
	}

	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestExportAsCurlNoBody(t *testing.T) {
	e := &store.Event{
		Method:  "GET",
		Path:    "/health",
		Headers: map[string][]string{},
	}

	out := ExportAsCurl(e)

	if strings.Contains(out, "-d") {
		t.Errorf("GET request should not have -d flag:\n%s", out)
	}
}

func TestExportAsCurlEscapesSingleQuotes(t *testing.T) {
	e := &store.Event{
		Method:  "POST",
		Path:    "/hook",
		Headers: map[string][]string{"X-Val": {"it's a test"}},
		Body:    []byte(`{"msg":"it's here"}`),
	}

	out := ExportAsCurl(e)

	if strings.Contains(out, "it's") {
		t.Errorf("single quotes not escaped:\n%s", out)
	}
}
