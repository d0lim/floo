package reverseproxy

import (
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
)

type NetHTTPProxy struct{}

func (p *NetHTTPProxy) Proxy(c *fiber.Ctx, upstream string) error {
	// Example: https://www.google.com + /google/hello = https://www.google.com/google/hello
	// Simple concatenation of Path (adjust as needed for actual use case)
	targetURL := fmt.Sprintf("%s%s", upstream, c.Request().URI().Path())

	// 1) Create a new HTTP request (including Body)
	req, err := http.NewRequest(c.Method(), targetURL, bytes.NewReader(c.Body()))
	if err != nil {
		return err
	}

	// 2) Copy request headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// 3) Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 4) Copy response to fiber.Ctx
	c.Status(resp.StatusCode)
	for k, vv := range resp.Header {
		for _, v := range vv {
			c.Append(k, v)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return c.Send(body)
}
