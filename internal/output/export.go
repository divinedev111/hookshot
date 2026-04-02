package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/divinedev111/hookshot/internal/store"
)

// ExportAsCurl converts a captured event into an executable curl command.
func ExportAsCurl(e *store.Event) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("curl -X %s", e.Method))

	path := e.Path
	if e.Query != "" {
		path += "?" + e.Query
	}
	parts = append(parts, fmt.Sprintf("  '%s'", "http://localhost"+path))

	keys := make([]string, 0, len(e.Headers))
	for k := range e.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		for _, v := range e.Headers[k] {
			parts = append(parts, fmt.Sprintf("  -H '%s: %s'", escapeQuotes(k), escapeQuotes(v)))
		}
	}

	if len(e.Body) > 0 {
		parts = append(parts, fmt.Sprintf("  -d '%s'", escapeQuotes(string(e.Body))))
	}

	return strings.Join(parts, " \\\n")
}

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}
