package reverseproxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// NetHTTPClient implements HTTPClient using the standard net/http package
type NetHTTPClient struct{}

// Execute performs an HTTP request using the net/http package
func (c *NetHTTPClient) Execute(method, url string, headers map[string][]string, body []byte) (int, map[string][]string, []byte, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}

	// Set headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, err
	}

	// Convert response headers
	respHeaders := make(map[string][]string)
	for k, v := range resp.Header {
		respHeaders[k] = v
	}

	return resp.StatusCode, respHeaders, respBody, nil
}

// NetHTTPProxy implements HTTPProxy interface using net/http package
type NetHTTPProxy struct {
	Client HTTPClient
}

// NewNetHTTPProxy creates a new NetHTTPProxy with the default NetHTTPClient
func NewNetHTTPProxy() *NetHTTPProxy {
	return &NetHTTPProxy{
		Client: &NetHTTPClient{},
	}
}

// Proxy implements the HTTPProxy interface
func (p *NetHTTPProxy) Proxy(c *fiber.Ctx, upstream string) error {
	// Construct target URL
	targetURL := fmt.Sprintf("%s%s", upstream, c.Request().URI().Path())

	// Extract headers from request
	headers := make(map[string][]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)
		if headers[k] == nil {
			headers[k] = []string{v}
		} else {
			headers[k] = append(headers[k], v)
		}
	})

	// Execute request through client implementation
	statusCode, respHeaders, respBody, err := p.Client.Execute(
		c.Method(),
		targetURL,
		headers,
		c.Body(),
	)
	if err != nil {
		return err
	}

	// Set response status
	c.Status(statusCode)

	// Set response headers
	for k, values := range respHeaders {
		for _, v := range values {
			c.Append(k, v)
		}
	}

	// Send response body
	return c.Send(respBody)
}
