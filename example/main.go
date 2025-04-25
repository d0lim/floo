package main

import (
	"fmt"
	"log"

	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// 1. 기본적인 net/http 기반 프록시 사용 예제
	netProxy := reverseproxy.NewNetHTTPProxy()
	app.Get("/net-proxy/*", func(c *fiber.Ctx) error {
		return netProxy.Proxy(c, "https://httpbin.org")
	})

	// 2. Fiber 클라이언트 기반 프록시 사용 예제
	fiberProxy := reverseproxy.NewFiberProxy()
	app.Get("/fiber-proxy/*", func(c *fiber.Ctx) error {
		return fiberProxy.Proxy(c, "https://httpbin.org")
	})

	// 3. HTTPClient 인터페이스를 활용한 커스텀 프록시 예제
	// 현실에서는 캐싱, 로드 밸런싱, 재시도 로직 등을 적용할 수 있음
	customClient := &reverseproxy.NetHTTPClient{} // 기본 HTTP 클라이언트 사용
	customProxy := &reverseproxy.NetHTTPProxy{
		Client: customClient,
	}
	app.Get("/custom-proxy/*", func(c *fiber.Ctx) error {
		return customProxy.Proxy(c, "https://httpbin.org")
	})

	port := 8080
	log.Printf("Server listening on port %d", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
