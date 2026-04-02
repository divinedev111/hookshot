package capture

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForward(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.Header.Get("X-Custom") != "test" {
			t.Errorf("missing custom header")
		}
		if string(body) != `{"test":true}` {
			t.Errorf("body = %q", body)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))
	defer srv.Close()

	headers := map[string][]string{"X-Custom": {"test"}}
	status, resp, err := Forward(srv.URL, "POST", headers, []byte(`{"test":true}`))
	if err != nil {
		t.Fatalf("forward: %v", err)
	}
	if status != 201 {
		t.Errorf("status = %d, want 201", status)
	}
	if string(resp) != "created" {
		t.Errorf("response = %q, want %q", resp, "created")
	}
}

func TestForwardAddsScheme(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Strip the http:// prefix to test auto-add
	addr := srv.URL[len("http://"):]
	status, _, err := Forward(addr, "GET", nil, nil)
	if err != nil {
		t.Fatalf("forward: %v", err)
	}
	if status != 200 {
		t.Errorf("status = %d, want 200", status)
	}
}
