package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/divinedev111/hookshot/internal/store"
)

var (
	providerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	idStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
)

// FormatEvent returns a single-line summary of an event.
func FormatEvent(e *store.Event, jsonMode bool) string {
	if jsonMode {
		return marshalJSON(eventSummary(e))
	}

	prov := "[unknown]"
	if e.Provider != "" {
		prov = fmt.Sprintf("[%s]", e.Provider)
	}

	evtType := ""
	if e.EventType != "" {
		evtType = e.EventType + "  "
	}

	fwd := ""
	if e.ForwardStatus != nil {
		fwd = fmt.Sprintf("-> %s  ", formatStatus(*e.ForwardStatus))
	}

	age := relativeTime(e.ReceivedAt)

	return fmt.Sprintf("%s  %s %s %s %s %s%s%s",
		idStyle.Render(fmt.Sprintf("#%d", e.ID)),
		providerStyle.Render(prov),
		evtType,
		e.Method,
		e.Path,
		fwd,
		dimStyle.Render(age),
		"",
	)
}

// FormatEventDetail returns full details for a single event.
func FormatEventDetail(e *store.Event, jsonMode bool) string {
	if jsonMode {
		return marshalJSON(eventDetail(e))
	}

	var b strings.Builder

	fmt.Fprintf(&b, "%s  %s\n", idStyle.Render(fmt.Sprintf("Event #%d", e.ID)), providerTag(e))
	fmt.Fprintf(&b, "  Method:     %s\n", e.Method)
	fmt.Fprintf(&b, "  Path:       %s\n", e.Path)
	if e.Query != "" {
		fmt.Fprintf(&b, "  Query:      %s\n", e.Query)
	}
	fmt.Fprintf(&b, "  Provider:   %s\n", providerName(e))
	if e.EventType != "" {
		fmt.Fprintf(&b, "  Event:      %s\n", e.EventType)
	}
	fmt.Fprintf(&b, "  Source IP:  %s\n", e.SourceIP)
	fmt.Fprintf(&b, "  Received:   %s\n", e.ReceivedAt.Format(time.RFC3339))

	if e.ForwardStatus != nil {
		fmt.Fprintf(&b, "  Forwarded:  %s\n", formatStatus(*e.ForwardStatus))
	}

	fmt.Fprintf(&b, "\n  Headers:\n")
	for k, vals := range e.Headers {
		for _, v := range vals {
			fmt.Fprintf(&b, "    %s: %s\n", k, v)
		}
	}

	if len(e.Body) > 0 {
		fmt.Fprintf(&b, "\n  Body:\n")
		fmt.Fprintf(&b, "    %s\n", prettyBody(e.Body))
	}

	if len(e.ForwardResponse) > 0 {
		fmt.Fprintf(&b, "\n  Forward Response:\n")
		fmt.Fprintf(&b, "    %s\n", prettyBody(e.ForwardResponse))
	}

	return b.String()
}

// FormatEventList returns a formatted table of events.
func FormatEventList(events []store.Event, jsonMode bool) string {
	if jsonMode {
		summaries := make([]map[string]any, len(events))
		for i := range events {
			summaries[i] = eventSummary(&events[i])
		}
		return marshalJSON(summaries)
	}

	if len(events) == 0 {
		return dimStyle.Render("No events found.")
	}

	var b strings.Builder
	for i, e := range events {
		b.WriteString(FormatEvent(&e, false))
		if i < len(events)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func formatStatus(code int) string {
	s := fmt.Sprintf("%d", code)
	if code >= 200 && code < 300 {
		return successStyle.Render(s)
	}
	return errorStyle.Render(s)
}

func providerTag(e *store.Event) string {
	if e.Provider == "" {
		return providerStyle.Render("[unknown]")
	}
	return providerStyle.Render(fmt.Sprintf("[%s]", e.Provider))
}

func providerName(e *store.Event) string {
	if e.Provider == "" {
		return "unknown"
	}
	return e.Provider
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func prettyBody(body []byte) string {
	var v any
	if json.Unmarshal(body, &v) == nil {
		b, _ := json.MarshalIndent(v, "    ", "  ")
		return string(b)
	}
	return string(body)
}

func eventSummary(e *store.Event) map[string]any {
	m := map[string]any{
		"id":          e.ID,
		"method":      e.Method,
		"path":        e.Path,
		"provider":    e.Provider,
		"event_type":  e.EventType,
		"received_at": e.ReceivedAt.Format(time.RFC3339),
	}
	if e.ForwardStatus != nil {
		m["forward_status"] = *e.ForwardStatus
	}
	return m
}

func eventDetail(e *store.Event) map[string]any {
	m := eventSummary(e)
	m["query"] = e.Query
	m["headers"] = e.Headers
	m["source_ip"] = e.SourceIP
	m["content_type"] = e.ContentType

	if len(e.Body) > 0 {
		var v any
		if json.Unmarshal(e.Body, &v) == nil {
			m["body"] = v
		} else {
			m["body"] = string(e.Body)
		}
	}

	if len(e.ForwardResponse) > 0 {
		var v any
		if json.Unmarshal(e.ForwardResponse, &v) == nil {
			m["forward_response"] = v
		} else {
			m["forward_response"] = string(e.ForwardResponse)
		}
	}

	return m
}

func marshalJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
