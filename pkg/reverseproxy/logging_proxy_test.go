package reverseproxy

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// 로그 캡처를 위한 버퍼
type logBuffer struct {
	buf bytes.Buffer
}

func (lb *logBuffer) Write(p []byte) (n int, err error) {
	return lb.buf.Write(p)
}

func (lb *logBuffer) String() string {
	return lb.buf.String()
}

func TestLoggingProxy(t *testing.T) {
	// 로그 캡처 설정
	logBuf := &logBuffer{}
	log.SetOutput(logBuf)
	log.SetFlags(0) // 타임스탬프 제거

	// 파이버 앱 생성
	app := fiber.New()

	// 목업 클라이언트 설정
	mockClient := &MockHTTPClient{
		StatusCode:  200,
		RespHeaders: map[string][]string{"Content-Type": {"application/json"}},
		RespBody:    []byte(`{"userId": 1, "id": 1, "title": "테스트 제목", "completed": false}`),
	}

	// 베이스 프록시
	baseProxy := &NetHTTPProxy{
		Client: mockClient,
	}

	// 로깅 프록시로 래핑
	loggingProxy := NewLoggingProxy(baseProxy)

	// 테스트 경로 추가
	app.Get("/test", func(c *fiber.Ctx) error {
		return loggingProxy.Proxy(c, "https://example.com")
	})

	// 테스트 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Test-Header", "테스트 값")

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

	// 필요한 로그 항목이 있는지 확인
	requiredLogItems := []string{
		"[요청] 매칭 경로: /test",
		"메서드: GET",
		"X-Test-Header",
		"[프록시] 대상 URL: https://example.com/test",
		"[응답] 상태 코드: 200",
		"[응답] 바디:",
	}

	for _, item := range requiredLogItems {
		if !strings.Contains(logs, item) {
			t.Errorf("로그에 '%s' 항목이 없습니다", item)
		}
	}
}

func TestLoggingProxyWithJSONPlaceholder(t *testing.T) {
	// 로그 캡처 설정
	logBuf := &logBuffer{}
	log.SetOutput(logBuf)
	log.SetFlags(0) // 타임스탬프 제거

	// 파이버 앱 생성
	app := fiber.New()

	// 베이스 프록시
	baseProxy := NewNetHTTPProxy()

	// 로깅 프록시로 래핑
	loggingProxy := NewLoggingProxy(baseProxy)

	// 테스트 경로 추가
	app.Get("/todos/:id", func(c *fiber.Ctx) error {
		return loggingProxy.Proxy(c, "https://jsonplaceholder.typicode.com")
	})

	// 테스트 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)

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

	// 필요한 로그 항목이 있는지 확인
	requiredLogItems := []string{
		"[요청] 매칭 경로: /todos/1",
		"메서드: GET",
		"[프록시] 대상 URL: https://jsonplaceholder.typicode.com/todos/1",
		"[응답] 상태 코드: 200",
		"delectus aut autem",
	}

	for _, item := range requiredLogItems {
		if !strings.Contains(logs, item) {
			t.Errorf("로그에 '%s' 항목이 없습니다", item)
		}
	}
}
