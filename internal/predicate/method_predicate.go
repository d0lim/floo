package predicate

import "github.com/gofiber/fiber/v2"

// MethodPredicate checks if the request method matches a specific method
type MethodPredicate struct {
	Method string
}

func (m MethodPredicate) Match(c *fiber.Ctx) bool {
	return c.Method() == m.Method
}
