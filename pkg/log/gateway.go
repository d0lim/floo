package log

import (
	"fmt"
	"time"

	"github.com/d0lim/floo/pkg/gateway"
	"github.com/gofiber/fiber/v2"
)

// GatewayLogger is a wrapper implementation that adds logging functionality to Gateway.
type GatewayLogger struct {
	Gateway gateway.Gateway
	Logger  Logger
}

// NewGatewayLogger creates a logging gateway that wraps an existing Gateway.
func NewGatewayLogger(gw gateway.Gateway) *GatewayLogger {
	return &GatewayLogger{
		Gateway: gw,
		Logger:  GetLogger(),
	}
}

// Handle wraps Gateway.Handle to add logging.
func (lg *GatewayLogger) Handle(c *fiber.Ctx) error {
	start := time.Now()
	logger := lg.Logger

	path := c.Path()
	method := c.Method()

	logger.Info(GatewayComponent, "Request received: path=%s, method=%s", path, method)

	// Try to match from defined Routes
	matchFound := false

	// Iterate through each route to check for a match
	for i, route := range lg.Gateway.Routes {
		routeStart := time.Now()

		// Log predicate matching
		logger.Debug(GatewayComponent, "Route[%d] matching started: %d predicates", i, len(route.Predicates))

		// Check each predicate individually
		allPredicatesMatched := true
		for j, pred := range route.Predicates {
			predMatch := pred.Match(c)
			logger.Debug(GatewayComponent, "  Predicate[%d]: %T matching result=%v", j, pred, predMatch)

			if !predMatch {
				allPredicatesMatched = false
				break
			}
		}

		// Check if all predicates matched
		if !allPredicatesMatched {
			logger.Debug(GatewayComponent, "Route[%d] matching failed: Predicate mismatch", i)
			continue
		}

		// Log matched route
		logger.Info(GatewayComponent, "Route[%d] matching successful: upstream=%s", i, route.Upstream)

		// Apply request filters
		if len(route.RequestFilters) > 0 {
			logger.Debug(GatewayComponent, "Applying request filters: %d filters", len(route.RequestFilters))

			for j, rf := range route.RequestFilters {
				filterDone := logger.Timed(FilterComponent, "Request filter[%d]: %T applying", j, rf)

				if err := rf.OnRequest(c); err != nil {
					logger.Error(FilterComponent, "Request filter[%d] application failed: %v", j, err)
					logger.Error(GatewayComponent, "Request processing failed: elapsed time=%s", time.Since(start))
					return err
				}

				filterDone("success")
			}
		}

		// Call the proxy if Upstream is set
		if route.Upstream != "" && lg.Gateway.ReverseProxy != nil {
			proxyDone := logger.Timed(ProxyComponent, "Proxy call: upstream=%s, path=%s", route.Upstream, c.Path())
			err := lg.Gateway.ReverseProxy.Proxy(c, route.Upstream)

			if err != nil {
				logger.Error(ProxyComponent, "Proxy call failed: %v", err)
				return err
			}

			proxyDone(fmt.Sprintf("success (status code=%d)", c.Response().StatusCode()))

			// Apply response filters
			if len(route.ResponseFilters) > 0 {
				logger.Debug(GatewayComponent, "Applying response filters: %d filters", len(route.ResponseFilters))

				for j, rf := range route.ResponseFilters {
					respFilterDone := logger.Timed(FilterComponent, "Response filter[%d]: %T applying", j, rf)

					if err := rf.OnResponse(c); err != nil {
						logger.Error(FilterComponent, "Response filter[%d] application failed: %v", j, err)
						return err
					}

					respFilterDone("success")
				}
			}

			matchFound = true
			routeElapsed := time.Since(routeStart)
			logger.Debug(GatewayComponent, "Route[%d] processing completed: elapsed time=%s", i, routeElapsed)
			break
		}
	}

	// Return 404 if no route matched
	if !matchFound {
		logger.Warn(GatewayComponent, "No matching route: returning 404")
		return fiber.NewError(fiber.StatusNotFound, "No matching route found")
	}

	elapsed := time.Since(start)
	logger.Info(GatewayComponent, "Request processing completed: path=%s, status=%d, elapsed time=%s",
		path, c.Response().StatusCode(), elapsed)

	return nil
}
