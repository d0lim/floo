# 리버스 프록시(Reverse Proxy) 시스템

이 패키지는 HTTP 요청을 다양한 백엔드 서비스로 프록시하는 기능을 제공합니다. 다형성(Polymorphism)을 활용하여 다양한 HTTP 클라이언트를 지원합니다.

## 특징

- 인터페이스 기반의 다형성 설계
- 기본 net/http 패키지 클라이언트 지원
- Fiber 클라이언트 지원
- 확장 가능한 아키텍처 (새로운 클라이언트 추가 용이)

## 주요 인터페이스와 클래스

### HTTPProxy

모든 프록시 구현체가 따라야 하는 인터페이스입니다.

```go
type HTTPProxy interface {
    Proxy(c *fiber.Ctx, upstream string) error
}
```

### HTTPClient

프록시 내부에서 사용되는 HTTP 클라이언트 인터페이스입니다.

```go
type HTTPClient interface {
    Execute(method, url string, headers map[string][]string, body []byte) (statusCode int, respHeaders map[string][]string, respBody []byte, err error)
}
```

### NetHTTPProxy

표준 라이브러리의 net/http 패키지를 사용하는 프록시 구현체입니다.

```go
proxy := reverseproxy.NewNetHTTPProxy()
```

### FiberProxy

Fiber의 HTTP 클라이언트를 사용하는 프록시 구현체입니다.

```go
proxy := reverseproxy.NewFiberProxy()
```

## 사용 예시

```go
app := fiber.New()

// 기본 net/http 클라이언트 사용 예제
netProxy := reverseproxy.NewNetHTTPProxy()
app.Get("/api/*", func(c *fiber.Ctx) error {
    return netProxy.Proxy(c, "https://api.example.com")
})

// Fiber 클라이언트 사용 예제
fiberProxy := reverseproxy.NewFiberProxy()
app.Get("/web/*", func(c *fiber.Ctx) error {
    return fiberProxy.Proxy(c, "https://web.example.com")
})

app.Listen(":3000")
```

## 커스텀 HTTP 클라이언트 구현 방법

자체 HTTP 클라이언트를 구현하려면 `HTTPClient` 인터페이스를 구현하세요:

```go
type MyCustomClient struct {
    // 커스텀 필드
}

func (c *MyCustomClient) Execute(method, url string, headers map[string][]string, body []byte) (int, map[string][]string, []byte, error) {
    // 커스텀 구현
    return statusCode, headers, body, nil
}

// 커스텀 클라이언트를 사용하는 프록시 생성
customProxy := &reverseproxy.NetHTTPProxy{
    Client: &MyCustomClient{},
}
```

## 테스트

기본 테스트 실행:

```bash
go test ./pkg/reverseproxy
``` 