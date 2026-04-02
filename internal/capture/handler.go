package capture

import (
	"io"
	"net/http"
	"time"

	"github.com/divinedev111/hookshot/internal/provider"
	"github.com/divinedev111/hookshot/internal/store"
)

// NewHandler returns an HTTP handler that captures incoming webhook requests.
func NewHandler(s *store.Store, forward string, onEvent func(*store.Event)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		info := provider.Detect(r.Header, body)

		e := &store.Event{
			Method:      r.Method,
			Path:        r.URL.Path,
			Query:       r.URL.RawQuery,
			Headers:     r.Header,
			Body:        body,
			ContentType: r.Header.Get("Content-Type"),
			Provider:    info.Name,
			EventType:   info.EventType,
			SourceIP:    r.RemoteAddr,
			ReceivedAt:  time.Now().UTC(),
		}

		if err := s.Save(e); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if forward != "" {
			status, resp, err := Forward(forward, e.Method, e.Headers, e.Body)
			if err == nil {
				s.UpdateForward(e.ID, status, resp)
				e.ForwardStatus = &status
				e.ForwardResponse = resp
			}
		}

		if onEvent != nil {
			onEvent(e)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}
