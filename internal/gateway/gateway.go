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
	// 정의해 둔 Routes 중 매칭되는 첫 번째 Route 처리
	for _, route := range g.Routes {
		if route.Match(c) {
			return route.Serve(c, g.ReverseProxy)
		}
	}
	// Return 404 when no matching route is found
	return fiber.NewError(http.StatusNotFound, "No matching route found")
}
