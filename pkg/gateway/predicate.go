package gateway

import "github.com/gofiber/fiber/v2"

// Predicate interface determines if a request matches a specific route
type Predicate interface {
	Match(*fiber.Ctx) bool
}
