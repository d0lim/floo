package gateway

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

// Route contains Predicates, Filters, and Upstream.
type Route struct {
	Predicates []Predicate
	Filters    []Filter
	Upstream   string // Target for Reverse Proxy (if empty, process locally)
}

// Match checks if this Route matches the current request.
func (r *Route) Match(c *fiber.Ctx) bool {
	for _, pred := range r.Predicates {
		if !pred.Match(c) {
			return false
		}
	}
	return true
}

// Serve applies filters, then processes locally or proxies to Upstream.
func (r *Route) Serve(c *fiber.Ctx, proxy ReverseProxy) error {
	// 1) Apply all Filters.
	for _, filter := range r.Filters {
		if err := filter.Apply(c); err != nil {
			return err
		}
	}

	// 2) Reverse Proxy if Upstream is set.
	if r.Upstream != "" && proxy != nil {
		return proxy.Proxy(c, r.Upstream)
	}

	// 3) Process locally
	return fiber.NewError(http.StatusNotFound, "No matching route found")
}
