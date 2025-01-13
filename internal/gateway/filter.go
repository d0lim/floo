package gateway

import "github.com/gofiber/fiber/v2"

// Filter is interface for pre/post processing of requests/responses
type Filter interface {
	Apply(*fiber.Ctx) error
}
