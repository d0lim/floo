package log

import (
	"fmt"
	"log"
	"time"

	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

// ProxyLogger는 다른 프록시를 감싸서 로깅 기능을 추가하는 래퍼 구현체입니다.
type ProxyLogger struct {
	Wrapped reverseproxy.HTTPProxy
}

// NewProxyLogger는 기존 프록시를 감싸는 로깅 프록시를 생성합니다.
func NewProxyLogger(wrapped reverseproxy.HTTPProxy) *ProxyLogger {
	return &ProxyLogger{
		Wrapped: wrapped,
	}
}

// Proxy는 HTTPProxy 인터페이스를 구현하고 로깅을 추가합니다.
func (p *ProxyLogger) Proxy(c *fiber.Ctx, upstream string) error {
	start := time.Now()

	// 요청 정보 로깅
	path := c.Path()
	method := c.Method()
	headers := map[string]string{}

	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	log.Printf("[요청] 매칭 경로: %s, 메서드: %s, 대상 업스트림: %s", path, method, upstream)
	log.Printf("[요청] 헤더: %v", headers)
	log.Printf("[요청] 바디: %s", string(c.Body()))

	// 타겟 URL 계산
	targetURL := fmt.Sprintf("%s%s", upstream, c.Request().URI().Path())
	log.Printf("[프록시] 대상 URL: %s", targetURL)

	// 원래 프록시 호출
	err := p.Wrapped.Proxy(c, upstream)

	// 응답 시간 계산
	elapsed := time.Since(start)

	// 응답 정보 로깅
	if err != nil {
		log.Printf("[응답] 오류 발생: %v", err)
	} else {
		log.Printf("[응답] 상태 코드: %d, 소요 시간: %s", c.Response().StatusCode(), elapsed)

		// 응답 헤더 로깅
		respHeaders := map[string]string{}
		c.Response().Header.VisitAll(func(key, value []byte) {
			respHeaders[string(key)] = string(value)
		})
		log.Printf("[응답] 헤더: %v", respHeaders)

		// 응답 바디 로깅 (너무 길 수 있으므로 일부만 로깅)
		respBody := c.Response().Body()
		if len(respBody) > 1024 {
			log.Printf("[응답] 바디 (처음 1KB): %s...", string(respBody[:1024]))
		} else {
			log.Printf("[응답] 바디: %s", string(respBody))
		}
	}

	return err
}
