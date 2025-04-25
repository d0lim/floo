package log

import (
	"fmt"

	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

// ProxyLogger is a wrapper implementation that adds logging functionality to another proxy.
type ProxyLogger struct {
	Wrapped reverseproxy.HTTPProxy
	Logger  Logger
}

// NewProxyLogger creates a logging proxy that wraps an existing proxy.
func NewProxyLogger(wrapped reverseproxy.HTTPProxy) *ProxyLogger {
	return &ProxyLogger{
		Wrapped: wrapped,
		Logger:  GetLogger(),
	}
}

// Proxy implements the HTTPProxy interface and adds logging.
func (p *ProxyLogger) Proxy(c *fiber.Ctx, upstream string) error {
	logger := p.Logger

	// Log request information
	path := c.Path()
	method := c.Method()
	headers := map[string]string{}

	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	logger.Info(ProxyComponent, "Request: path=%s, method=%s, target=%s", path, method, upstream)

	if IsDebugEnabled() {
		logger.Debug(ProxyComponent, "Request headers: %v", headers)
		logger.Debug(ProxyComponent, "Request body: %s", string(c.Body()))
	}

	// Calculate target URL
	targetURL := fmt.Sprintf("%s%s", upstream, c.Request().URI().Path())
	logger.Debug(ProxyComponent, "Target URL: %s", targetURL)

	// Call the original proxy
	proxyDone := logger.Timed(ProxyComponent, "Sending proxy request: %s %s", method, targetURL)
	err := p.Wrapped.Proxy(c, upstream)

	if err != nil {
		logger.Error(ProxyComponent, "Error occurred: %v", err)
		return err
	}

	statusCode := c.Response().StatusCode()
	proxyDone(fmt.Sprintf("Response received: status=%d", statusCode))

	// Detailed logging only in debug mode
	if IsDebugEnabled() {
		// Log response headers
		respHeaders := map[string]string{}
		c.Response().Header.VisitAll(func(key, value []byte) {
			respHeaders[string(key)] = string(value)
		})
		logger.Debug(ProxyComponent, "Response headers: %v", respHeaders)

		// Log response body (only partial if too long)
		respBody := c.Response().Body()
		if len(respBody) > 1024 {
			logger.Debug(ProxyComponent, "Response body (first 1KB): %s...", string(respBody[:1024]))
		} else {
			logger.Debug(ProxyComponent, "Response body: %s", string(respBody))
		}
	}

	return nil
}
