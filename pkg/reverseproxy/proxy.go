package reverseproxy

import (
	"github.com/gofiber/fiber/v2"
)

// HTTPProxy defines an interface for various HTTP proxy implementations
// All proxy implementations must implement this interface
type HTTPProxy interface {
	// Proxy sends the request to the upstream and copies the response to fiber.Ctx
	Proxy(c *fiber.Ctx, upstream string) error
}

// HTTPClient is an interface for HTTP clients used by proxies
// It abstracts the HTTP client implementation, allowing different client libraries
type HTTPClient interface {
	// Execute performs an HTTP request and returns the response
	Execute(method, url string, headers map[string][]string, body []byte) (statusCode int, respHeaders map[string][]string, respBody []byte, err error)
}
