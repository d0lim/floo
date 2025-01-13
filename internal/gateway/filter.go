package gateway

import "github.com/gofiber/fiber/v2"

// RequestFilter only involves in pre-processing "requests"
type RequestFilter interface {
	OnRequest(c *fiber.Ctx) error
}

// ResponseFilter only involves in post-processing "responses"
type ResponseFilter interface {
	OnResponse(c *fiber.Ctx) error
}
