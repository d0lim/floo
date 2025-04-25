package reverseproxy

import (
	"github.com/gofiber/fiber/v2"
)

// FiberHTTPClient implements HTTPClient using Fiber's client package
type FiberHTTPClient struct {
	agent *fiber.Agent
}

// NewFiberHTTPClient creates a new FiberHTTPClient
func NewFiberHTTPClient() *FiberHTTPClient {
	return &FiberHTTPClient{
		agent: fiber.AcquireAgent(),
	}
}

// Execute performs an HTTP request using the Fiber client
func (c *FiberHTTPClient) Execute(method, url string, headers map[string][]string, body []byte) (int, map[string][]string, []byte, error) {
	// Create a reusable agent
	agent := c.agent.Reuse()

	// Set URL and method
	req := agent.Request()
	req.Header.SetMethod(method)
	req.SetRequestURI(url)

	// Set body if provided
	if len(body) > 0 {
		agent.Body(body)
	}

	// Set headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Create a custom response to get headers
	resp := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(resp)
	agent.SetResponse(resp)

	// Execute request
	if err := agent.Parse(); err != nil {
		return 0, nil, nil, err
	}

	statusCode, respBody, errs := agent.Bytes()
	if len(errs) > 0 {
		return 0, nil, nil, errs[0]
	}

	// Get response headers
	respHeaders := make(map[string][]string)
	resp.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)
		if respHeaders[k] == nil {
			respHeaders[k] = []string{v}
		} else {
			respHeaders[k] = append(respHeaders[k], v)
		}
	})

	return statusCode, respHeaders, respBody, nil
}

// FiberProxy implements HTTPProxy interface using Fiber's client package
type FiberProxy struct {
	Client HTTPClient
}

// NewFiberProxy creates a new FiberProxy with the default FiberHTTPClient
func NewFiberProxy() *FiberProxy {
	return &FiberProxy{
		Client: NewFiberHTTPClient(),
	}
}

// Proxy implements the HTTPProxy interface
func (p *FiberProxy) Proxy(c *fiber.Ctx, upstream string) error {
	// Construct target URL
	targetURL := upstream + string(c.Request().URI().Path())

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
