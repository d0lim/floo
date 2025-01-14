package filter

import "github.com/gofiber/fiber/v2"

type AddHeaderResponseFilter struct {
	Key   string
	Value string
}

func (f AddHeaderResponseFilter) OnResponse(c *fiber.Ctx) error {
	c.Set(f.Key, f.Value)
	return nil
}
