# Reverse Proxy System

This package provides functionality for proxying HTTP requests to various backend services. It utilizes polymorphism to support different HTTP clients.

## Features

- Interface-based polymorphic design
- Support for standard net/http package client
- Support for Fiber client
- Extensible architecture (easy to add new clients)

## Main Interfaces and Classes

### HTTPProxy

The interface that all proxy implementations must follow.

```go
type HTTPProxy interface {
    Proxy(c *fiber.Ctx, upstream string) error
}
```

### HTTPClient

The HTTP client interface used within proxies.

```go
type HTTPClient interface {
    Execute(method, url string, headers map[string][]string, body []byte) (statusCode int, respHeaders map[string][]string, respBody []byte, err error)
}
```

### NetHTTPProxy

A proxy implementation using the standard library's net/http package.

```go
proxy := reverseproxy.NewNetHTTPProxy()
```

### FiberProxy

A proxy implementation using Fiber's HTTP client.

```go
proxy := reverseproxy.NewFiberProxy()
```

## Usage Examples

```go
app := fiber.New()

// Example using the default net/http client
netProxy := reverseproxy.NewNetHTTPProxy()
app.Get("/api/*", func(c *fiber.Ctx) error {
    return netProxy.Proxy(c, "https://api.example.com")
})

// Example using the Fiber client
fiberProxy := reverseproxy.NewFiberProxy()
app.Get("/web/*", func(c *fiber.Ctx) error {
    return fiberProxy.Proxy(c, "https://web.example.com")
})

app.Listen(":3000")
```

## How to Implement a Custom HTTP Client

To implement your own HTTP client, implement the `HTTPClient` interface:

```go
type MyCustomClient struct {
    // Custom fields
}

func (c *MyCustomClient) Execute(method, url string, headers map[string][]string, body []byte) (int, map[string][]string, []byte, error) {
    // Custom implementation
    return statusCode, headers, body, nil
}

// Create a proxy that uses the custom client
customProxy := &reverseproxy.NetHTTPProxy{
    Client: &MyCustomClient{},
}
```

## Testing

Run basic tests:

```bash
go test ./pkg/reverseproxy
``` 