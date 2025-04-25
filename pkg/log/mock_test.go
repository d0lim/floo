package log

import (
	"github.com/d0lim/floo/pkg/reverseproxy"
)

// MockHTTPClient is an HTTPClient implementation for testing.
type MockHTTPClient struct {
	StatusCode  int
	RespHeaders map[string][]string
	RespBody    []byte
	Error       error
}

// Execute returns predefined response values.
func (m *MockHTTPClient) Execute(method, url string, headers map[string][]string, body []byte) (int, map[string][]string, []byte, error) {
	return m.StatusCode, m.RespHeaders, m.RespBody, m.Error
}

// Verify that MockHTTPClient satisfies the HTTPClient interface (compile-time check)
var _ reverseproxy.HTTPClient = (*MockHTTPClient)(nil)
