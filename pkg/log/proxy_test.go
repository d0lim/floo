package log

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

// TodoResponse is a struct for jsonplaceholder Todo response
type TodoResponse struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func TestProxyLogger(t *testing.T) {
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
		RespBody:    []byte(`{"userId": 1, "id": 1, "title": "Test Title", "completed": false}`),
	}

	// Base proxy
	baseProxy := &reverseproxy.NetHTTPProxy{
		Client: mockClient,
	}

	// Wrap with logging proxy
	loggingProxy := NewProxyLogger(baseProxy)

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return loggingProxy.Proxy(c, "https://example.com")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Test-Header", "Test Value")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request test failed: %v", err)
	}

	// Check response
	if resp.StatusCode != 200 {
		t.Errorf("Status code should be 200, but got %d", resp.StatusCode)
	}

	// Parse JSON response
	var todo TodoResponse
	err = json.NewDecoder(resp.Body).Decode(&todo)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Validate expected response fields
	if todo.UserID != 1 {
		t.Errorf("UserID should be 1, but got %d", todo.UserID)
	}
	if todo.ID != 1 {
		t.Errorf("ID should be 1, but got %d", todo.ID)
	}
	if todo.Title != "Test Title" {
		t.Errorf("Title should be 'Test Title', but got '%s'", todo.Title)
	}
	if todo.Completed != false {
		t.Errorf("Completed should be false, but got %t", todo.Completed)
	}

	// Check logs
	logs := logBuf.String()
	t.Logf("Log output: %s", logs)

	// Verify required log items are present - adjusted for new log format
	requiredLogItems := []string{
		"[Proxy][INFO] Request: path=/test",
		"method=GET",
		"[Proxy][DEBUG] Request headers",
		"X-Test-Header",
		"[Proxy][DEBUG] Target URL: https://example.com/test",
		"[Proxy][INFO] Sending proxy request: GET https://example.com/test",
		"Response received: status=200",
		"[Proxy][DEBUG] Response body",
	}

	for _, item := range requiredLogItems {
		if !strings.Contains(logs, item) {
			t.Errorf("Log does not contain '%s' item", item)
		}
	}
}

func TestProxyLoggerWithJSONPlaceholder(t *testing.T) {
	// Use Mock instead of calling actual API
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

	// Mock client that simulates JSONPlaceholder API response
	mockClient := &MockHTTPClient{
		StatusCode: 200,
		RespHeaders: map[string][]string{
			"Content-Type": {"application/json"},
		},
		RespBody: []byte(`{"userId": 1, "id": 1, "title": "delectus aut autem", "completed": false}`),
	}

	// Create base proxy with Mock client
	baseProxy := &reverseproxy.NetHTTPProxy{
		Client: mockClient,
	}

	// Wrap with logging proxy
	loggingProxy := NewProxyLogger(baseProxy)

	// Add test route
	app.Get("/todos/:id", func(c *fiber.Ctx) error {
		return loggingProxy.Proxy(c, "https://jsonplaceholder.typicode.com")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request test failed: %v", err)
	}

	// Check response
	if resp.StatusCode != 200 {
		t.Errorf("Status code should be 200, but got %d", resp.StatusCode)
	}

	// Parse JSON response
	var todo TodoResponse
	err = json.NewDecoder(resp.Body).Decode(&todo)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Validate expected response fields
	if todo.UserID != 1 {
		t.Errorf("UserID should be 1, but got %d", todo.UserID)
	}
	if todo.ID != 1 {
		t.Errorf("ID should be 1, but got %d", todo.ID)
	}
	if todo.Title != "delectus aut autem" {
		t.Errorf("Title should be 'delectus aut autem', but got '%s'", todo.Title)
	}
	if todo.Completed != false {
		t.Errorf("Completed should be false, but got %t", todo.Completed)
	}

	// Check logs
	logs := logBuf.String()
	t.Logf("Log output: %s", logs)

	// Verify required log items are present - adjusted for new log format
	requiredLogItems := []string{
		"[Proxy][INFO] Request: path=/todos/1",
		"method=GET",
		"[Proxy][DEBUG] Target URL: https://jsonplaceholder.typicode.com/todos/1",
		"[Proxy][INFO] Sending proxy request",
		"Response received: status=200",
		"delectus aut autem",
	}

	for _, item := range requiredLogItems {
		if !strings.Contains(logs, item) {
			t.Errorf("Log does not contain '%s' item", item)
		}
	}
}
