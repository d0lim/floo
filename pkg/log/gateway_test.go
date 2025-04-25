package log

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/d0lim/floo/pkg/gateway"
	"github.com/d0lim/floo/pkg/predicate"
	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

// MockPredicate는 항상 지정된 결과를 반환하는 Predicate 구현체입니다.
type MockPredicate struct {
	Result bool
}

// Match는 저장된 Result 값을 반환합니다.
func (p MockPredicate) Match(c *fiber.Ctx) bool {
	return p.Result
}

// MockRequestFilter는 요청 필터 구현체입니다.
type MockRequestFilter struct {
	Error error
}

// OnRequest는 저장된 Error 값을 반환합니다.
func (f MockRequestFilter) OnRequest(c *fiber.Ctx) error {
	return f.Error
}

func TestGatewayLogger(t *testing.T) {
	// 로그 캡처 설정
	logBuf := NewBuffer()
	restore := CaptureLogsToBuffer(logBuf)
	defer restore()

	// 기본 로그 설정 초기화
	ConfigureLogger(LogFlags{}, "") // 타임스탬프 제거

	// 테스트를 위해 디버그 레벨 활성화
	SetLogLevel(DebugLevel)

	// 파이버 앱 생성
	app := fiber.New()

	// 목업 클라이언트 설정
	mockClient := &MockHTTPClient{
		StatusCode:  200,
		RespHeaders: map[string][]string{"Content-Type": {"application/json"}},
		RespBody:    []byte(`{"success": true}`),
	}

	// 베이스 프록시
	baseProxy := &reverseproxy.NetHTTPProxy{
		Client: mockClient,
	}

	// 기본 게이트웨이 생성
	baseGateway := gateway.Gateway{
		ReverseProxy: baseProxy,
		Routes: []gateway.Route{
			{
				Predicates: []gateway.Predicate{
					predicate.PathPrefixPredicate{Prefix: "/api"},
					MockPredicate{Result: true}, // 항상 매칭되는 Predicate
				},
				RequestFilters: []gateway.RequestFilter{
					MockRequestFilter{Error: nil}, // 항상 성공하는 필터
				},
				Upstream: "https://example.com",
			},
			{
				Predicates: []gateway.Predicate{
					MockPredicate{Result: false}, // 절대 매칭되지 않는 Predicate
				},
				Upstream: "https://never-matched.com",
			},
		},
	}

	// 로깅 게이트웨이로 래핑
	loggingGateway := NewGatewayLogger(baseGateway)

	// 테스트 경로 추가
	app.All("/*", loggingGateway.Handle)

	// 테스트 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)

	// 요청 실행
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("요청 테스트 실패: %v", err)
	}

	// 응답 확인
	if resp.StatusCode != 200 {
		t.Errorf("상태 코드가 200이어야 하는데, %d를 받았습니다", resp.StatusCode)
	}

	// 로그 확인
	logs := logBuf.String()
	t.Logf("로그 출력: %s", logs)

	// 필요한 로그 항목이 있는지 확인 - 새로운 로그 형식에 맞춤
	requiredLogItems := []string{
		"[게이트웨이][INFO] 요청 수신: 경로=/api/test",
		"[게이트웨이][DEBUG] 라우트[0] 매칭 시작",
		"Predicate[0]",
		"Predicate[1]",
		"[게이트웨이][INFO] 라우트[0] 매칭 성공",
		"[필터][INFO] 요청 필터[0]",
		"[프록시][INFO] 프록시 호출",
		"성공 (상태 코드=200)",
		"[게이트웨이][INFO] 요청 처리 완료",
	}

	for _, item := range requiredLogItems {
		if !strings.Contains(logs, item) {
			t.Errorf("로그에 '%s' 항목이 없습니다", item)
		}
	}

	// 404 경로에 대한 테스트
	req = httptest.NewRequest(http.MethodGet, "/not-found", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 404 {
		t.Errorf("없는 경로에 대해 상태 코드가 404여야 하는데, %d를 받았습니다", resp.StatusCode)
	}
}
