package predicate

import "github.com/gofiber/fiber/v2"

// PathPredicate checks if the request path exactly matches a specific path
type PathPredicate struct {
	Path string
}

func (p PathPredicate) Match(c *fiber.Ctx) bool {
	return c.Path() == p.Path
}

// PathPrefixPredicate checks if the request path starts with a specific prefix
type PathPrefixPredicate struct {
	Prefix string
}

func (p PathPrefixPredicate) Match(c *fiber.Ctx) bool {
	return len(c.Path()) >= len(p.Prefix) && c.Path()[:len(p.Prefix)] == p.Prefix
}
