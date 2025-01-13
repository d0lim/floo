package filter

import "github.com/gofiber/fiber/v2"

// RewritePathFilter is a Filter that rewrites the path
type RewritePathFilter struct {
	From string
	To   string
}

func (f RewritePathFilter) Apply(c *fiber.Ctx) error {
	if c.Path() == f.From {
		c.Request().URI().SetPath(f.To)
	}
	return nil
}
