package main

import (
	"fmt"
	"github.com/d0lim/floo/pkg/filter"
	"github.com/d0lim/floo/pkg/gateway"
	"github.com/d0lim/floo/pkg/predicate"
	"github.com/d0lim/floo/pkg/reverseproxy"
	"log"
	"regexp"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	p := &reverseproxy.NetHTTPProxy{}

	gw := gateway.Gateway{
		ReverseProxy: p,
		Routes: []gateway.Route{
			{
				Predicates: []gateway.Predicate{
					predicate.PathPrefixPredicate{Prefix: "/placeholder"},
				},
				RequestFilters: []gateway.RequestFilter{
					filter.AddHeaderRequestFilter{Key: "X-Proxy", Value: "Go-Floo-Gateway"},
					filter.RewritePathRequestFilter{Pattern: regexp.MustCompile(`^/placeholder/(.*)`),
						Replacement: "/$1"},
				},
				Upstream: "https://jsonplaceholder.typicode.com",
			},
		},
	}

	app.Get("/api/v1/ping", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.All("/*", gw.Handle)

	port := 8080
	log.Printf("Gateway listening on port %d\n", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
