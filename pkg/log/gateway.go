package log

import (
	"fmt"
	"time"

	"github.com/d0lim/floo/pkg/gateway"
	"github.com/gofiber/fiber/v2"
)

// GatewayLogger는 Gateway에 로깅 기능을 추가한 래퍼 구현체입니다.
type GatewayLogger struct {
	Gateway gateway.Gateway
	Logger  Logger
}

// NewGatewayLogger는 기존 Gateway를 감싸는 로깅 게이트웨이를 생성합니다.
func NewGatewayLogger(gw gateway.Gateway) *GatewayLogger {
	return &GatewayLogger{
		Gateway: gw,
		Logger:  GetLogger(),
	}
}

// Handle은 Gateway.Handle을 래핑하여 로깅을 추가합니다.
func (lg *GatewayLogger) Handle(c *fiber.Ctx) error {
	start := time.Now()
	logger := lg.Logger

	path := c.Path()
	method := c.Method()

	logger.Info(GatewayComponent, "요청 수신: 경로=%s, 메서드=%s", path, method)

	// 정의해 둔 Routes 중 매칭 시도
	matchFound := false

	// 각 라우트를 순회하며 매칭 확인
	for i, route := range lg.Gateway.Routes {
		routeStart := time.Now()

		// Predicate 매칭 로깅
		logger.Debug(GatewayComponent, "라우트[%d] 매칭 시작: %d개 predicate", i, len(route.Predicates))

		// 각 Predicate 개별 확인
		allPredicatesMatched := true
		for j, pred := range route.Predicates {
			predMatch := pred.Match(c)
			logger.Debug(GatewayComponent, "  Predicate[%d]: %T 매칭 결과=%v", j, pred, predMatch)

			if !predMatch {
				allPredicatesMatched = false
				break
			}
		}

		// 모든 Predicate가 매치되었는지 확인
		if !allPredicatesMatched {
			logger.Debug(GatewayComponent, "라우트[%d] 매칭 실패: Predicate 불일치", i)
			continue
		}

		// 매칭된 라우트 로깅
		logger.Info(GatewayComponent, "라우트[%d] 매칭 성공: 업스트림=%s", i, route.Upstream)

		// 요청 필터 적용
		if len(route.RequestFilters) > 0 {
			logger.Debug(GatewayComponent, "요청 필터 적용: %d개", len(route.RequestFilters))

			for j, rf := range route.RequestFilters {
				filterDone := logger.Timed(FilterComponent, "요청 필터[%d]: %T 적용", j, rf)

				if err := rf.OnRequest(c); err != nil {
					logger.Error(FilterComponent, "요청 필터[%d] 적용 실패: %v", j, err)
					logger.Error(GatewayComponent, "요청 처리 실패: 소요 시간=%s", time.Since(start))
					return err
				}

				filterDone("성공")
			}
		}

		// Upstream이 설정되어 있으면 프록시 호출
		if route.Upstream != "" && lg.Gateway.ReverseProxy != nil {
			proxyDone := logger.Timed(ProxyComponent, "프록시 호출: 업스트림=%s, 경로=%s", route.Upstream, c.Path())
			err := lg.Gateway.ReverseProxy.Proxy(c, route.Upstream)

			if err != nil {
				logger.Error(ProxyComponent, "프록시 호출 실패: %v", err)
				return err
			}

			proxyDone(fmt.Sprintf("성공 (상태 코드=%d)", c.Response().StatusCode()))

			// 응답 필터 적용
			if len(route.ResponseFilters) > 0 {
				logger.Debug(GatewayComponent, "응답 필터 적용: %d개", len(route.ResponseFilters))

				for j, rf := range route.ResponseFilters {
					respFilterDone := logger.Timed(FilterComponent, "응답 필터[%d]: %T 적용", j, rf)

					if err := rf.OnResponse(c); err != nil {
						logger.Error(FilterComponent, "응답 필터[%d] 적용 실패: %v", j, err)
						return err
					}

					respFilterDone("성공")
				}
			}

			matchFound = true
			routeElapsed := time.Since(routeStart)
			logger.Debug(GatewayComponent, "라우트[%d] 처리 완료: 소요 시간=%s", i, routeElapsed)
			break
		}
	}

	// 매칭된 라우트가 없으면 404 반환
	if !matchFound {
		logger.Warn(GatewayComponent, "매칭 라우트 없음: 404 반환")
		return fiber.NewError(fiber.StatusNotFound, "No matching route found")
	}

	elapsed := time.Since(start)
	logger.Info(GatewayComponent, "요청 처리 완료: 경로=%s, 상태=%d, 소요 시간=%s",
		path, c.Response().StatusCode(), elapsed)

	return nil
}
