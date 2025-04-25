package main

import (
	"fmt"
	"regexp"

	"github.com/d0lim/floo/pkg/filter"
	"github.com/d0lim/floo/pkg/gateway"
	"github.com/d0lim/floo/pkg/log"
	"github.com/d0lim/floo/pkg/predicate"
	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// 로그 설정
	log.ConfigureLogger(log.LogFlags{Time: true, File: true}, "[FLOO] ")
	log.SetLogLevel(log.DebugLevel) // 디버그 모드 활성화
	logger := log.GetLogger()

	logger.Info(log.GatewayComponent, "전체 로깅 예제 애플리케이션을 시작합니다...")

	app := fiber.New()

	// 기본 프록시 생성 및 로깅 프록시로 래핑
	baseProxy := reverseproxy.NewNetHTTPProxy()
	loggingProxy := log.NewProxyLogger(baseProxy)

	// 기본 게이트웨이 생성
	baseGateway := gateway.Gateway{
		ReverseProxy: loggingProxy,
		Routes: []gateway.Route{
			{
				Predicates: []gateway.Predicate{
					predicate.PathPrefixPredicate{Prefix: "/todos"},
					predicate.MethodPredicate{Method: "GET"},
				},
				RequestFilters: []gateway.RequestFilter{
					filter.AddHeaderRequestFilter{Key: "X-Proxy", Value: "Go-Floo-Gateway"},
				},
				Upstream: "https://jsonplaceholder.typicode.com",
			},
			{
				Predicates: []gateway.Predicate{
					predicate.PathPrefixPredicate{Prefix: "/posts"},
				},
				RequestFilters: []gateway.RequestFilter{
					filter.AddHeaderRequestFilter{Key: "X-Proxy", Value: "Go-Floo-Gateway"},
				},
				Upstream: "https://jsonplaceholder.typicode.com",
			},
			{
				Predicates: []gateway.Predicate{
					predicate.PathPrefixPredicate{Prefix: "/echo"},
				},
				RequestFilters: []gateway.RequestFilter{
					filter.AddHeaderRequestFilter{Key: "X-Echo-Test", Value: "로깅 예제"},
					filter.RewritePathRequestFilter{
						Pattern:     regexp.MustCompile(`^/echo/(.*)`),
						Replacement: "/$1",
					},
				},
				Upstream: "https://postman-echo.com",
			},
		},
	}

	// 로깅 게이트웨이로 래핑
	loggingGateway := log.NewGatewayLogger(baseGateway)

	// 테스트용 핑 엔드포인트
	app.Get("/api/ping", func(c *fiber.Ctx) error {
		logger.Debug(log.GatewayComponent, "핑 요청 수신")
		return c.SendString("OK")
	})

	// 모든 다른 경로는 게이트웨이로 라우팅
	app.All("/*", loggingGateway.Handle)

	port := 8083
	logger.Info(log.GatewayComponent, "전체 로깅 게이트웨이가 포트 %d에서 시작됩니다", port)
	logger.Info(log.GatewayComponent, "테스트 URL 예시:")
	logger.Info(log.GatewayComponent, "  - http://localhost:%d/todos/1        (GET만 허용)", port)
	logger.Info(log.GatewayComponent, "  - http://localhost:%d/posts/1        (모든 메서드 허용)", port)
	logger.Info(log.GatewayComponent, "  - http://localhost:%d/echo/get?foo=bar", port)
	app.Listen(fmt.Sprintf(":%d", port))
}
