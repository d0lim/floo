package main

import (
	"fmt"
	"regexp"

	"github.com/d0lim/floo/pkg/filter"
	"github.com/d0lim/floo/pkg/gateway"
	"github.com/d0lim/floo/pkg/log"
	"github.com/d0lim/floo/pkg/predicate"
	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Log configuration
	log.ConfigureLogger(log.LogFlags{Time: true, File: true}, "[FLOO] ")
	log.SetLogLevel(log.DebugLevel) // Enable debug mode
	logger := log.GetLogger()

	logger.Info(log.GatewayComponent, "Starting full logging example application...")

	app := fiber.New()

	// Create base proxy and wrap with logging proxy
	baseProxy := reverseproxy.NewNetHTTPProxy()
	loggingProxy := log.NewProxyLogger(baseProxy)

	// Create base gateway
	baseGateway := gateway.Gateway{
		ReverseProxy: loggingProxy,
		Routes: []gateway.Route{
			{
				Predicates: []gateway.Predicate{
					predicate.PathPrefixPredicate{Prefix: "/todos"},
					predicate.MethodPredicate{Method: "GET"},
				},
				RequestFilters: []gateway.RequestFilter{
					filter.AddHeaderRequestFilter{Key: "X-Proxy", Value: "Go-Floo-Gateway"},
				},
				Upstream: "https://jsonplaceholder.typicode.com",
			},
			{
				Predicates: []gateway.Predicate{
					predicate.PathPrefixPredicate{Prefix: "/posts"},
				},
				RequestFilters: []gateway.RequestFilter{
					filter.AddHeaderRequestFilter{Key: "X-Proxy", Value: "Go-Floo-Gateway"},
				},
				Upstream: "https://jsonplaceholder.typicode.com",
			},
			{
				Predicates: []gateway.Predicate{
					predicate.PathPrefixPredicate{Prefix: "/echo"},
				},
				RequestFilters: []gateway.RequestFilter{
					filter.AddHeaderRequestFilter{Key: "X-Echo-Test", Value: "Logging Example"},
					filter.RewritePathRequestFilter{
						Pattern:     regexp.MustCompile(`^/echo/(.*)`),
						Replacement: "/$1",
					},
				},
				Upstream: "https://postman-echo.com",
			},
		},
	}

	// Wrap with logging gateway
	loggingGateway := log.NewGatewayLogger(baseGateway)

	// Test ping endpoint
	app.Get("/api/ping", func(c *fiber.Ctx) error {
		logger.Debug(log.GatewayComponent, "Ping request received")
		return c.SendString("OK")
	})

	// Route all other paths to the gateway
	app.All("/*", loggingGateway.Handle)

	port := 8083
	logger.Info(log.GatewayComponent, "Full logging gateway starting on port %d", port)
	logger.Info(log.GatewayComponent, "Test URL examples:")
	logger.Info(log.GatewayComponent, "  - http://localhost:%d/todos/1        (GET only)", port)
	logger.Info(log.GatewayComponent, "  - http://localhost:%d/posts/1        (All methods allowed)", port)
	logger.Info(log.GatewayComponent, "  - http://localhost:%d/echo/get?foo=bar", port)
	app.Listen(fmt.Sprintf(":%d", port))
}
