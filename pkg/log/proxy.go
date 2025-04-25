package log

import (
	"fmt"

	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

// ProxyLogger는 다른 프록시를 감싸서 로깅 기능을 추가하는 래퍼 구현체입니다.
type ProxyLogger struct {
	Wrapped reverseproxy.HTTPProxy
	Logger  Logger
}

// NewProxyLogger는 기존 프록시를 감싸는 로깅 프록시를 생성합니다.
func NewProxyLogger(wrapped reverseproxy.HTTPProxy) *ProxyLogger {
	return &ProxyLogger{
		Wrapped: wrapped,
		Logger:  GetLogger(),
	}
}

// Proxy는 HTTPProxy 인터페이스를 구현하고 로깅을 추가합니다.
func (p *ProxyLogger) Proxy(c *fiber.Ctx, upstream string) error {
	logger := p.Logger

	// 요청 정보 로깅
	path := c.Path()
	method := c.Method()
	headers := map[string]string{}

	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	logger.Info(ProxyComponent, "요청: 경로=%s, 메서드=%s, 대상=%s", path, method, upstream)

	if IsDebugEnabled() {
		logger.Debug(ProxyComponent, "요청 헤더: %v", headers)
		logger.Debug(ProxyComponent, "요청 바디: %s", string(c.Body()))
	}

	// 타겟 URL 계산
	targetURL := fmt.Sprintf("%s%s", upstream, c.Request().URI().Path())
	logger.Debug(ProxyComponent, "대상 URL: %s", targetURL)

	// 원래 프록시 호출
	proxyDone := logger.Timed(ProxyComponent, "프록시 요청 전송: %s %s", method, targetURL)
	err := p.Wrapped.Proxy(c, upstream)

	if err != nil {
		logger.Error(ProxyComponent, "오류 발생: %v", err)
		return err
	}

	statusCode := c.Response().StatusCode()
	proxyDone(fmt.Sprintf("응답 수신: 상태=%d", statusCode))

	// 디버그 모드에서만 상세 로깅
	if IsDebugEnabled() {
		// 응답 헤더 로깅
		respHeaders := map[string]string{}
		c.Response().Header.VisitAll(func(key, value []byte) {
			respHeaders[string(key)] = string(value)
		})
		logger.Debug(ProxyComponent, "응답 헤더: %v", respHeaders)

		// 응답 바디 로깅 (너무 길 수 있으므로 일부만 로깅)
		respBody := c.Response().Body()
		if len(respBody) > 1024 {
			logger.Debug(ProxyComponent, "응답 바디 (처음 1KB): %s...", string(respBody[:1024]))
		} else {
			logger.Debug(ProxyComponent, "응답 바디: %s", string(respBody))
		}
	}

	return nil
}
