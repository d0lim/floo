package gateway

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LoggingGateway는 Gateway에 로깅 기능을 추가한 래퍼 구현체입니다.
type LoggingGateway struct {
	Gateway Gateway
}

// NewLoggingGateway는 기존 Gateway를 감싸는 로깅 게이트웨이를 생성합니다.
func NewLoggingGateway(gw Gateway) *LoggingGateway {
	return &LoggingGateway{
		Gateway: gw,
	}
}

// Handle은 Gateway.Handle을 래핑하여 로깅을 추가합니다.
func (lg *LoggingGateway) Handle(c *fiber.Ctx) error {
	start := time.Now()

	path := c.Path()
	method := c.Method()

	log.Printf("[게이트웨이] 요청 수신: 경로=%s, 메서드=%s", path, method)

	// 정의해 둔 Routes 중 매칭 시도
	matchFound := false

	// 각 라우트를 순회하며 매칭 확인
	for i, route := range lg.Gateway.Routes {
		routeStart := time.Now()

		// Predicate 매칭 로깅
		log.Printf("[게이트웨이] 라우트[%d] 매칭 시작: %d개 predicate", i, len(route.Predicates))

		// 각 Predicate 개별 확인
		allPredicatesMatched := true
		for j, pred := range route.Predicates {
			predMatch := pred.Match(c)
			log.Printf("[게이트웨이]   Predicate[%d]: %T 매칭 결과=%v", j, pred, predMatch)

			if !predMatch {
				allPredicatesMatched = false
				break
			}
		}

		// 모든 Predicate가 매치되었는지 확인
		if !allPredicatesMatched {
			log.Printf("[게이트웨이] 라우트[%d] 매칭 실패: Predicate 불일치", i)
			continue
		}

		// 매칭된 라우트 로깅
		log.Printf("[게이트웨이] 라우트[%d] 매칭 성공: 업스트림=%s", i, route.Upstream)
		log.Printf("[게이트웨이] 요청 필터 적용: %d개", len(route.RequestFilters))

		// 요청 필터 적용
		for j, rf := range route.RequestFilters {
			filterStart := time.Now()

			log.Printf("[게이트웨이]   요청 필터[%d]: %T 적용 시작", j, rf)
			if err := rf.OnRequest(c); err != nil {
				log.Printf("[게이트웨이]   요청 필터[%d]: 적용 실패: %v", j, err)
				elapsed := time.Since(start)
				log.Printf("[게이트웨이] 요청 처리 완료: 소요 시간=%s, 오류 발생", elapsed)
				return err
			}

			filterElapsed := time.Since(filterStart)
			log.Printf("[게이트웨이]   요청 필터[%d]: 적용 완료: 소요 시간=%s", j, filterElapsed)
		}

		// Upstream이 설정되어 있으면 프록시 호출
		if route.Upstream != "" && lg.Gateway.ReverseProxy != nil {
			proxyStart := time.Now()

			log.Printf("[게이트웨이] 프록시 호출: 업스트림=%s, 경로=%s", route.Upstream, c.Path())
			err := lg.Gateway.ReverseProxy.Proxy(c, route.Upstream)

			proxyElapsed := time.Since(proxyStart)
			if err != nil {
				log.Printf("[게이트웨이] 프록시 호출 실패: %v, 소요 시간=%s", err, proxyElapsed)
				return err
			}

			log.Printf("[게이트웨이] 프록시 호출 성공: 상태 코드=%d, 소요 시간=%s", c.Response().StatusCode(), proxyElapsed)

			// 응답 필터 적용
			if len(route.ResponseFilters) > 0 {
				log.Printf("[게이트웨이] 응답 필터 적용: %d개", len(route.ResponseFilters))

				for j, rf := range route.ResponseFilters {
					respFilterStart := time.Now()

					log.Printf("[게이트웨이]   응답 필터[%d]: %T 적용 시작", j, rf)
					if err := rf.OnResponse(c); err != nil {
						log.Printf("[게이트웨이]   응답 필터[%d]: 적용 실패: %v", j, err)
						return err
					}

					respFilterElapsed := time.Since(respFilterStart)
					log.Printf("[게이트웨이]   응답 필터[%d]: 적용 완료: 소요 시간=%s", j, respFilterElapsed)
				}
			}

			matchFound = true
			routeElapsed := time.Since(routeStart)
			log.Printf("[게이트웨이] 라우트[%d] 처리 완료: 소요 시간=%s", i, routeElapsed)
			break
		}
	}

	// 매칭된 라우트가 없으면 404 반환
	if !matchFound {
		log.Printf("[게이트웨이] 매칭 라우트 없음: 404 반환")
		return fiber.NewError(fiber.StatusNotFound, "No matching route found")
	}

	elapsed := time.Since(start)
	log.Printf("[게이트웨이] 요청 처리 완료: 경로=%s, 상태=%d, 소요 시간=%s",
		path, c.Response().StatusCode(), elapsed)

	return nil
}
