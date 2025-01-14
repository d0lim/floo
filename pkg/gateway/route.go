package gateway

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

// Route contains Predicates, Filters, and Upstream.
type Route struct {
	Predicates      []Predicate
	RequestFilters  []RequestFilter
	ResponseFilters []ResponseFilter
	Upstream        string
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

// Serve applies filters, then processes proxies to Upstream.
func (r *Route) Serve(c *fiber.Ctx, proxy ReverseProxy) error {
	// 1) Apply all RequestFilters
	for _, rf := range r.RequestFilters {
		if err := rf.OnRequest(c); err != nil {
			return err
		}
	}

	// 2) Reverse Proxy if Upstream is set
	if r.Upstream != "" && proxy != nil {
		return proxy.Proxy(c, r.Upstream)
	}

	// 3) Apply all ResponseFilters
	for _, rf := range r.ResponseFilters {
		if err := rf.OnResponse(c); err != nil {
			return err
		}
	}

	// 4) Return 404 when no matching route is found
	return fiber.NewError(http.StatusNotFound, "No matching route found")
}
