package provider

// Info holds the detected provider name and event type.
type Info struct {
	Name      string
	EventType string
}

type definition struct {
	name       string
	header     string
	eventType  func(headers map[string][]string, body []byte) string
}

var providers = []definition{
	{
		name:   "stripe",
		header: "Stripe-Signature",
		eventType: func(_ map[string][]string, body []byte) string {
			return jsonField(body, "type")
		},
	},
	{
		name:   "github",
		header: "X-Github-Event",
		eventType: func(h map[string][]string, _ []byte) string {
			return firstHeader(h, "X-Github-Event")
		},
	},
	{
		name:   "shopify",
		header: "X-Shopify-Hmac-Sha256",
		eventType: func(h map[string][]string, _ []byte) string {
			return firstHeader(h, "X-Shopify-Topic")
		},
	},
	{
		name:   "slack",
		header: "X-Slack-Signature",
		eventType: func(_ map[string][]string, body []byte) string {
			return jsonField(body, "type")
		},
	},
	{
		name:   "discord",
		header: "X-Signature-Ed25519",
		eventType: func(_ map[string][]string, body []byte) string {
			return jsonField(body, "type")
		},
	},
	{
		name:   "twilio",
		header: "X-Twilio-Signature",
		eventType: func(_ map[string][]string, body []byte) string {
			return formField(body, "EventType")
		},
	},
}
