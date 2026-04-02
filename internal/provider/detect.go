package provider

import (
	"encoding/json"
	"net/url"
	"strings"
)

// Detect identifies the webhook provider and event type from request headers and body.
func Detect(headers map[string][]string, body []byte) Info {
	for _, p := range providers {
		if hasHeader(headers, p.header) {
			return Info{
				Name:      p.name,
				EventType: p.eventType(headers, body),
			}
		}
	}
	return Info{}
}

func hasHeader(h map[string][]string, key string) bool {
	key = strings.ToLower(key)
	for k, v := range h {
		if strings.ToLower(k) == key && len(v) > 0 && v[0] != "" {
			return true
		}
	}
	return false
}

func firstHeader(h map[string][]string, key string) string {
	key = strings.ToLower(key)
	for k, v := range h {
		if strings.ToLower(k) == key && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

func jsonField(body []byte, field string) string {
	var m map[string]any
	if json.Unmarshal(body, &m) != nil {
		return ""
	}
	if v, ok := m[field].(string); ok {
		return v
	}
	return ""
}

func formField(body []byte, field string) string {
	vals, err := url.ParseQuery(string(body))
	if err != nil {
		return ""
	}
	return vals.Get(field)
}
