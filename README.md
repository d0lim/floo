# Floo

[![Go Version](https://img.shields.io/github/go-mod/go-version/gofiber/fiber?style=flat-square)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](./LICENSE)

**Floo** is a **Go + Fiber**-based **Gateway** framework, inspired by the magical **Floo Powder** from Harry Potter, known for instant teleportation.  
Like wizards traveling between fireplaces, Floo aims to teleport your requests to the correct **Upstream** services quickly and transparently.

---

## Overview

Floo is designed to function as a **routing** and **reverse proxy** layer for multiple upstream services. All requests are routed according to custom **Predicates** and transformed by **Filters** before being proxied to an **Upstream**.

> **Important**: Unlike some gateway examples that allow “local responses,” Floo requires each **Route** to have a **non-optional Upstream**. This ensures that your gateway is purely responsible for **teleporting** (proxying) requests and does not mix business logic.

---

## Key Features

- **Go + Fiber**: Built on top of Fiber’s performant and easy-to-use HTTP framework.
- **Strict Reverse Proxy**: Each Route must define a non-empty **Upstream**, ensuring consistent external communication.
- **Predicate-based Routing**: Match requests by paths, prefixes, methods, or any custom logic.
- **Request/Response Filtering**: Transform or inspect requests and responses before or after passing them through.
- **Extensible**: Implement custom Predicates, Filters, or even different Proxy clients.
- **Inspired by Spring Cloud Gateway**: But with the magical twist of “Floo travel” for your requests.

---

## Architecture

### Route

A **Route** is defined by:

1. **Predicates**: Conditions to determine if the incoming request should be handled by this Route.
2. **Request Filters**: Pre-processing logic (e.g., rewriting paths, adding headers).
3. **Upstream**: The **non-optional** target service URL where the request is ultimately sent.
4. **Response Filters**: Post-processing logic (e.g., modifying response headers, logging).

### Predicates

Predicates determine if a Route matches a given request. Examples include:

- **PathPredicate**: Matches an exact path.
- **PathPrefixPredicate**: Matches paths that begin with a given prefix.
- **MethodPredicate**: Matches a specific HTTP method (GET, POST, etc.).

### Request Filters (WIP)

**Request Filters** (`RequestFilter`) operate on the incoming request **before** it’s sent to the Upstream. Typical use cases:

- Adding or modifying headers (e.g., correlation IDs, auth tokens).
- Rewriting path segments (e.g., removing `"/api"` prefix).
- Logging inbound request details.

### Response Filters (WIP)

**Response Filters** (`ResponseFilter`) run **after** receiving a response from the Upstream. Potential operations include:

- Injecting or modifying response headers.
- Transforming the response body (e.g., masking sensitive data).
- Logging or metrics collection on response details.

### Reverse Proxy

Floo includes a pluggable **Reverse Proxy** component.
- It receives the prepared request (after Request Filters) and sends it to the specified `Upstream`.
- It captures the response (status, headers, body) and applies all **Response Filters** before finalizing the response to the client.

---

## Getting Started

### Installation

```bash
go get github.com/gofiber/fiber/v2
go get github.com/d0lim/floo
```

### Usage

1. **Define** a list of **Routes**, each with: Predicates, Filters, and a mandatory Upstream.
2. **Initialize** a Floo `Gateway` with a chosen Reverse Proxy implementation (e.g., `NetHTTPProxy`).
3. **Attach** Floo’s main handler (`Gateway.Handle`) to your Fiber application.
4. **Start** the Fiber server.

---

## Example

```go
package main

import (
	"fmt"
	"github.com/d0lim/floo/internal/filter"
	"github.com/d0lim/floo/internal/gateway"
	"github.com/d0lim/floo/internal/predicate"
	"github.com/d0lim/floo/internal/reverseproxy"
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
```

### Testing

```bash
# Forward request to JSONPlaceholder
curl http://localhost:8080/placeholder/todos/1
```

---

## Roadmap

- **Advanced Predicates**: Host-based, header-based, or advanced regex routes.
- **Custom Filters**: Authentication, Rate Limiting, Circuit Breaker, Observability.
- **Dynamic Configuration**: Manage routes via database or remote config.
- **Plugin Architecture**: Allow user-defined plugin modules for more specialized transformations.

---

## Contributing

Contributions are welcome and encouraged! Feel free to:

1. **Fork** this repository
2. **Create** a new branch for your feature or fix (`git checkout -b feature/my-feature`)
3. **Commit** and **push** your changes
4. **Open** a Pull Request detailing your modifications

Please make sure to include tests or examples, and ensure that existing functionality is not broken.

---

## License

This project is under the [MIT License](./LICENSE).  
Feel free to use and adapt Floo in your own projects, and consider contributing back any improvements you make!

---

**May your requests travel quickly and safely — with just a pinch of Floo magic!**