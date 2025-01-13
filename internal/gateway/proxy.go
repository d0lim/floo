package gateway

import "github.com/gofiber/fiber/v2"

// ReverseProxy Interface
// - Send a request to a specific Upstream and copy the result to fiber.Ctx
type ReverseProxy interface {
	Proxy(ctx *fiber.Ctx, upstream string) error
}
