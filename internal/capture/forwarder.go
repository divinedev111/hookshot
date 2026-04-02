package capture

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var forwardClient = &http.Client{Timeout: 30 * time.Second}

// Forward sends a request with the given method, headers, and body to the target URL.
func Forward(target, method string, headers map[string][]string, body []byte) (status int, respBody []byte, err error) {
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "http://" + target
	}

	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return 0, nil, fmt.Errorf("creating forward request: %w", err)
	}

	for k, vals := range headers {
		if strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	resp, err := forwardClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("forwarding request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("reading forward response: %w", err)
	}

	return resp.StatusCode, respBody, nil
}
