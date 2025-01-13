package filter

import "github.com/gofiber/fiber/v2"

// AddHeaderFilter is a simple filter that adds response headers
type AddHeaderFilter struct {
	Key   string
	Value string
}

func (f AddHeaderFilter) Apply(c *fiber.Ctx) error {
	c.Set(f.Key, f.Value)
	return nil
}
