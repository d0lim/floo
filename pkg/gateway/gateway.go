package gateway

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Gateway contains multiple Routes and appropriately routes incoming requests
type Gateway struct {
	Routes       []Route
	ReverseProxy ReverseProxy
}

// Handle is handler of Fiber
func (g *Gateway) Handle(c *fiber.Ctx) error {
	// Process the first matching Route from the defined Routes
	for _, route := range g.Routes {
		if route.Match(c) {
			return route.Serve(c, g.ReverseProxy)
		}
	}
	// Return 404 when no matching route is found
	return fiber.NewError(http.StatusNotFound, "No matching route found")
}
