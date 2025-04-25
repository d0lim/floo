package reverseproxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TodoResponse는 jsonplaceholder의 Todo 응답 구조체
type TodoResponse struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func TestNetHTTPProxyWithJSONPlaceholder(t *testing.T) {
	// 파이버 앱 생성
	app := fiber.New()

	// 실제 NetHTTPProxy 생성 (실제 HTTP 요청 사용)
	proxy := NewNetHTTPProxy()

	// 테스트 경로 추가
	app.Get("/todos/:id", func(c *fiber.Ctx) error {
		return proxy.Proxy(c, "https://jsonplaceholder.typicode.com")
	})

	// 테스트 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("요청 테스트 실패: %v", err)
	}

	// 상태 코드 확인
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
}

func TestNetHTTPProxyWithMockJSONPlaceholder(t *testing.T) {
	// 파이버 앱 생성
	app := fiber.New()

	// 목업 응답 준비
	mockClient := &MockHTTPClient{
		StatusCode: 200,
		RespHeaders: map[string][]string{
			"Content-Type": {"application/json"},
		},
		RespBody: []byte(`{"userId": 1, "id": 1, "title": "delectus aut autem", "completed": false}`),
	}

	// 목업 클라이언트로 프록시 생성
	proxy := &NetHTTPProxy{
		Client: mockClient,
	}

	// 테스트 경로 추가
	app.Get("/todos/:id", func(c *fiber.Ctx) error {
		return proxy.Proxy(c, "https://jsonplaceholder.typicode.com")
	})

	// 테스트 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("요청 테스트 실패: %v", err)
	}

	// 상태 코드 확인
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
}
