package filter

import (
	"github.com/gofiber/fiber/v2"
	"regexp"
)

type AddHeaderRequestFilter struct {
	Key   string
	Value string
}

func (f AddHeaderRequestFilter) OnRequest(c *fiber.Ctx) error {
	c.Set(f.Key, f.Value)
	return nil
}

type RewritePathRequestFilter struct {
	Pattern     *regexp.Regexp
	Replacement string
}

func (f RewritePathRequestFilter) OnRequest(c *fiber.Ctx) error {
	originalPath := c.Path()
	newPath := f.Pattern.ReplaceAllString(originalPath, f.Replacement)
	c.Request().URI().SetPath(newPath)
	return nil
}
