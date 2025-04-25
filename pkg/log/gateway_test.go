package log

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/d0lim/floo/pkg/gateway"
	"github.com/d0lim/floo/pkg/predicate"
	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

// MockPredicate is a Predicate implementation that always returns the specified result.
type MockPredicate struct {
	Result bool
}

// Match returns the stored Result value.
func (p MockPredicate) Match(c *fiber.Ctx) bool {
	return p.Result
}

// MockRequestFilter is a request filter implementation.
type MockRequestFilter struct {
	Error error
}

// OnRequest returns the stored Error value.
func (f MockRequestFilter) OnRequest(c *fiber.Ctx) error {
	return f.Error
}

func TestGatewayLogger(t *testing.T) {
	// Setup log capture
	logBuf := NewBuffer()
	restore := CaptureLogsToBuffer(logBuf)
	defer restore()

	// Initialize basic log configuration
	ConfigureLogger(LogFlags{}, "") // Remove timestamp

	// Enable debug level for testing
	SetLogLevel(DebugLevel)

	// Create fiber app
	app := fiber.New()

	// Setup mock client
	mockClient := &MockHTTPClient{
		StatusCode:  200,
		RespHeaders: map[string][]string{"Content-Type": {"application/json"}},
		RespBody:    []byte(`{"success": true}`),
	}

	// Base proxy
	baseProxy := &reverseproxy.NetHTTPProxy{
		Client: mockClient,
	}

	// Create base gateway
	baseGateway := gateway.Gateway{
		ReverseProxy: baseProxy,
		Routes: []gateway.Route{
			{
				Predicates: []gateway.Predicate{
					predicate.PathPrefixPredicate{Prefix: "/api"},
					MockPredicate{Result: true}, // Predicate that always matches
				},
				RequestFilters: []gateway.RequestFilter{
					MockRequestFilter{Error: nil}, // Filter that always succeeds
				},
				Upstream: "https://example.com",
			},
			{
				Predicates: []gateway.Predicate{
					MockPredicate{Result: false}, // Predicate that never matches
				},
				Upstream: "https://never-matched.com",
			},
		},
	}

	// Wrap with logging gateway
	loggingGateway := NewGatewayLogger(baseGateway)

	// Add test route
	app.All("/*", loggingGateway.Handle)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request test failed: %v", err)
	}

	// Check response
	if resp.StatusCode != 200 {
		t.Errorf("Status code should be 200, but got %d", resp.StatusCode)
	}

	// Check logs
	logs := logBuf.String()
	t.Logf("Log output: %s", logs)

	// Verify required log items are present - adjusted for new log format
	requiredLogItems := []string{
		"[Gateway][INFO] Request received: path=/api/test",
		"[Gateway][DEBUG] Route[0] matching started",
		"Predicate[0]",
		"Predicate[1]",
		"[Gateway][INFO] Route[0] matching successful",
		"[Filter][INFO] Request filter[0]",
		"[Proxy][INFO] Proxy call",
		"success (status code=200)",
		"[Gateway][INFO] Request processing completed",
	}

	for _, item := range requiredLogItems {
		if !strings.Contains(logs, item) {
			t.Errorf("Log does not contain '%s' item", item)
		}
	}

	// Test for 404 path
	req = httptest.NewRequest(http.MethodGet, "/not-found", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 404 {
		t.Errorf("Status code should be 404 for non-existent path, but got %d", resp.StatusCode)
	}
}
