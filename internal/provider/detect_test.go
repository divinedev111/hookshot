package provider

import "testing"

func TestDetect(t *testing.T) {
	tests := []struct {
		name      string
		headers   map[string][]string
		body      []byte
		wantName  string
		wantEvent string
	}{
		{
			name:      "stripe",
			headers:   map[string][]string{"Stripe-Signature": {"t=123,v1=abc"}},
			body:      []byte(`{"type":"payment_intent.succeeded"}`),
			wantName:  "stripe",
			wantEvent: "payment_intent.succeeded",
		},
		{
			name:      "github",
			headers:   map[string][]string{"X-Github-Event": {"push"}},
			body:      nil,
			wantName:  "github",
			wantEvent: "push",
		},
		{
			name:      "shopify",
			headers:   map[string][]string{"X-Shopify-Hmac-Sha256": {"abc"}, "X-Shopify-Topic": {"orders/create"}},
			body:      nil,
			wantName:  "shopify",
			wantEvent: "orders/create",
		},
		{
			name:      "slack",
			headers:   map[string][]string{"X-Slack-Signature": {"v0=abc"}},
			body:      []byte(`{"type":"event_callback"}`),
			wantName:  "slack",
			wantEvent: "event_callback",
		},
		{
			name:      "discord",
			headers:   map[string][]string{"X-Signature-Ed25519": {"abc"}},
			body:      []byte(`{"type":"1"}`),
			wantName:  "discord",
			wantEvent: "1",
		},
		{
			name:      "twilio",
			headers:   map[string][]string{"X-Twilio-Signature": {"abc"}},
			body:      []byte(`EventType=incoming&From=%2B1234`),
			wantName:  "twilio",
			wantEvent: "incoming",
		},
		{
			name:      "unknown",
			headers:   map[string][]string{"Content-Type": {"application/json"}},
			body:      []byte(`{"foo":"bar"}`),
			wantName:  "",
			wantEvent: "",
		},
		{
			name:      "case insensitive headers",
			headers:   map[string][]string{"stripe-signature": {"t=123"}},
			body:      []byte(`{"type":"charge.created"}`),
			wantName:  "stripe",
			wantEvent: "charge.created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := Detect(tt.headers, tt.body)
			if info.Name != tt.wantName {
				t.Errorf("name = %q, want %q", info.Name, tt.wantName)
			}
			if info.EventType != tt.wantEvent {
				t.Errorf("event_type = %q, want %q", info.EventType, tt.wantEvent)
			}
		})
	}
}

func TestDetectWithInvalidBody(t *testing.T) {
	info := Detect(
		map[string][]string{"Stripe-Signature": {"t=123"}},
		[]byte(`not json`),
	)
	if info.Name != "stripe" {
		t.Errorf("name = %q, want stripe", info.Name)
	}
	if info.EventType != "" {
		t.Errorf("event_type = %q, want empty", info.EventType)
	}
}
