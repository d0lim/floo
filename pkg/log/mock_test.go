package log

import (
	"github.com/d0lim/floo/pkg/reverseproxy"
)

// MockHTTPClient는 테스트를 위한 HTTPClient 구현체입니다.
type MockHTTPClient struct {
	StatusCode  int
	RespHeaders map[string][]string
	RespBody    []byte
	Error       error
}

// Execute는 사전 정의된 응답값을 반환합니다.
func (m *MockHTTPClient) Execute(method, url string, headers map[string][]string, body []byte) (int, map[string][]string, []byte, error) {
	return m.StatusCode, m.RespHeaders, m.RespBody, m.Error
}

// HTTPProxy 인터페이스를 만족하는지 확인 (컴파일 타임 체크)
var _ reverseproxy.HTTPClient = (*MockHTTPClient)(nil)
