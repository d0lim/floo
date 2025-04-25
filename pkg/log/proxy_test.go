package log

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/d0lim/floo/pkg/reverseproxy"
	"github.com/gofiber/fiber/v2"
)

// TodoResponse는 jsonplaceholder의 Todo 응답 구조체
type TodoResponse struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func TestProxyLogger(t *testing.T) {
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
		RespBody:    []byte(`{"userId": 1, "id": 1, "title": "테스트 제목", "completed": false}`),
	}

	// 베이스 프록시
	baseProxy := &reverseproxy.NetHTTPProxy{
		Client: mockClient,
	}

	// 로깅 프록시로 래핑
	loggingProxy := NewProxyLogger(baseProxy)

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

	// JSON 응답 파싱
	var todo TodoResponse
	err = json.NewDecoder(resp.Body).Decode(&todo)
	if err != nil {
		t.Fatalf("JSON 응답 파싱 실패: %v", err)
	}

	// 예상 응답 필드 검증
	if todo.UserID != 1 {
		t.Errorf("UserID는 1이어야 하는데, %d를 받았습니다", todo.UserID)
	}
	if todo.ID != 1 {
		t.Errorf("ID는 1이어야 하는데, %d를 받았습니다", todo.ID)
	}
	if todo.Title != "테스트 제목" {
		t.Errorf("제목이 '테스트 제목'이어야 하는데, '%s'를 받았습니다", todo.Title)
	}
	if todo.Completed != false {
		t.Errorf("Completed는 false여야 하는데, %t를 받았습니다", todo.Completed)
	}

	// 로그 확인
	logs := logBuf.String()
	t.Logf("로그 출력: %s", logs)

	// 필요한 로그 항목이 있는지 확인 - 새로운 로그 형식에 맞춤
	requiredLogItems := []string{
		"[프록시][INFO] 요청: 경로=/test",
		"메서드=GET",
		"[프록시][DEBUG] 요청 헤더",
		"X-Test-Header",
		"[프록시][DEBUG] 대상 URL: https://example.com/test",
		"[프록시][INFO] 프록시 요청 전송: GET https://example.com/test",
		"응답 수신: 상태=200",
		"[프록시][DEBUG] 응답 바디",
	}

	for _, item := range requiredLogItems {
		if !strings.Contains(logs, item) {
			t.Errorf("로그에 '%s' 항목이 없습니다", item)
		}
	}
}

func TestProxyLoggerWithJSONPlaceholder(t *testing.T) {
	// 실제 API를 호출하는 대신 Mock을 사용
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

	// JSONPlaceholder API 응답을 시뮬레이션하는 목업 클라이언트
	mockClient := &MockHTTPClient{
		StatusCode: 200,
		RespHeaders: map[string][]string{
			"Content-Type": {"application/json"},
		},
		RespBody: []byte(`{"userId": 1, "id": 1, "title": "delectus aut autem", "completed": false}`),
	}

	// 베이스 프록시를 Mock 클라이언트와 함께 생성
	baseProxy := &reverseproxy.NetHTTPProxy{
		Client: mockClient,
	}

	// 로깅 프록시로 래핑
	loggingProxy := NewProxyLogger(baseProxy)

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

	// JSON 응답 파싱
	var todo TodoResponse
	err = json.NewDecoder(resp.Body).Decode(&todo)
	if err != nil {
		t.Fatalf("JSON 응답 파싱 실패: %v", err)
	}

	// 예상 응답 필드 검증
	if todo.UserID != 1 {
		t.Errorf("UserID는 1이어야 하는데, %d를 받았습니다", todo.UserID)
	}
	if todo.ID != 1 {
		t.Errorf("ID는 1이어야 하는데, %d를 받았습니다", todo.ID)
	}
	if todo.Title != "delectus aut autem" {
		t.Errorf("제목이 'delectus aut autem'이어야 하는데, '%s'를 받았습니다", todo.Title)
	}
	if todo.Completed != false {
		t.Errorf("Completed는 false여야 하는데, %t를 받았습니다", todo.Completed)
	}

	// 로그 확인
	logs := logBuf.String()
	t.Logf("로그 출력: %s", logs)

	// 필요한 로그 항목이 있는지 확인 - 새로운 로그 형식에 맞춤
	requiredLogItems := []string{
		"[프록시][INFO] 요청: 경로=/todos/1",
		"메서드=GET",
		"[프록시][DEBUG] 대상 URL: https://jsonplaceholder.typicode.com/todos/1",
		"[프록시][INFO] 프록시 요청 전송",
		"응답 수신: 상태=200",
		"delectus aut autem",
	}

	for _, item := range requiredLogItems {
		if !strings.Contains(logs, item) {
			t.Errorf("로그에 '%s' 항목이 없습니다", item)
		}
	}
}
