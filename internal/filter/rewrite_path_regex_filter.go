package filter

import (
	"fmt"
	"regexp"

	"github.com/gofiber/fiber/v2"
)

// RewritePathRegexFilter filters that rewrites the request path using a regular expression
type RewritePathRegexFilter struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// Apply rewrites the path using a regular expression in the pre-request stage
func (f RewritePathRegexFilter) Apply(c *fiber.Ctx) error {
	originalPath := c.Path()
	newPath := f.Pattern.ReplaceAllString(originalPath, f.Replacement)
	fmt.Println("Rewriting path from", originalPath, "to", newPath)
	c.Request().URI().SetPath(newPath)
	return nil
}
