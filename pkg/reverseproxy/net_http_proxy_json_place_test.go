package reverseproxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TodoResponse is a struct for jsonplaceholder Todo response
type TodoResponse struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func TestNetHTTPProxyWithJSONPlaceholder(t *testing.T) {
	// Create fiber app
	app := fiber.New()

	// Create actual NetHTTPProxy (using real HTTP requests)
	proxy := NewNetHTTPProxy()

	// Add test route
	app.Get("/todos/:id", func(c *fiber.Ctx) error {
		return proxy.Proxy(c, "https://jsonplaceholder.typicode.com")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request test failed: %v", err)
	}

	// Check status code
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
}

func TestNetHTTPProxyWithMockJSONPlaceholder(t *testing.T) {
	// Create fiber app
	app := fiber.New()

	// Prepare mock response
	mockClient := &MockHTTPClient{
		StatusCode: 200,
		RespHeaders: map[string][]string{
			"Content-Type": {"application/json"},
		},
		RespBody: []byte(`{"userId": 1, "id": 1, "title": "delectus aut autem", "completed": false}`),
	}

	// Create proxy with mock client
	proxy := &NetHTTPProxy{
		Client: mockClient,
	}

	// Add test route
	app.Get("/todos/:id", func(c *fiber.Ctx) error {
		return proxy.Proxy(c, "https://jsonplaceholder.typicode.com")
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request test failed: %v", err)
	}

	// Check status code
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
}
