package log

import (
	"bytes"
	"fmt"
	"log"
	"time"
)

// LogLevel은 로그 레벨 상수를 정의합니다.
type LogLevel int

const (
	// DebugLevel은 디버깅용 상세 로그를 위한 레벨입니다.
	DebugLevel LogLevel = iota
	// InfoLevel은 정보 제공용 로그를 위한 레벨입니다.
	InfoLevel
	// WarnLevel은 경고 로그를 위한 레벨입니다.
	WarnLevel
	// ErrorLevel은 오류 로그를 위한 레벨입니다.
	ErrorLevel
)

// ComponentType은 로깅에 사용되는 컴포넌트 유형입니다.
type ComponentType string

const (
	// GatewayComponent는 게이트웨이 컴포넌트를 나타냅니다.
	GatewayComponent ComponentType = "게이트웨이"
	// ProxyComponent는 프록시 컴포넌트를 나타냅니다.
	ProxyComponent ComponentType = "프록시"
	// FilterComponent는 필터 컴포넌트를 나타냅니다.
	FilterComponent ComponentType = "필터"
	// PredicateComponent는 조건부 컴포넌트를 나타냅니다.
	PredicateComponent ComponentType = "조건부"
)

// LogFlags는 로그 출력 형식을 정의하는 구조체입니다.
type LogFlags struct {
	// 날짜 포함 여부
	Date bool
	// 시간 포함 여부
	Time bool
	// 마이크로초 포함 여부
	Microseconds bool
	// UTC 시간 사용 여부
	UTC bool
	// 파일 이름 포함 여부
	File bool
	// 전체 파일 경로 사용 여부
	LongFile bool
}

// Logger는 Floo의 로깅을 위한 공통 인터페이스입니다.
type Logger interface {
	Debug(component ComponentType, format string, v ...interface{})
	Info(component ComponentType, format string, v ...interface{})
	Warn(component ComponentType, format string, v ...interface{})
	Error(component ComponentType, format string, v ...interface{})
	Timed(component ComponentType, format string, v ...interface{}) func(result string)
}

// StandardLogger는 표준 로깅 구현체입니다.
type StandardLogger struct {
	logger *log.Logger
	level  LogLevel
}

var (
	// 현재 로그 레벨, 기본값은 InfoLevel
	currentLevel = InfoLevel
	// 기본 로거
	defaultLogger = log.Default()
	// 공유 로거 인스턴스
	sharedLogger Logger = &StandardLogger{logger: defaultLogger, level: currentLevel}
)

// SetLogLevel은 현재 로그 레벨을 설정합니다.
func SetLogLevel(level LogLevel) {
	currentLevel = level
	if stdLogger, ok := sharedLogger.(*StandardLogger); ok {
		stdLogger.level = level
	}
}

// GetLogLevel은 현재 로그 레벨을 반환합니다.
func GetLogLevel() LogLevel {
	return currentLevel
}

// IsDebugEnabled는 디버그 로그가 활성화되었는지 확인합니다.
func IsDebugEnabled() bool {
	return currentLevel <= DebugLevel
}

// GetLogger는 공유 로거를 반환합니다.
func GetLogger() Logger {
	return sharedLogger
}

// Debug는 디버그 레벨 로그를 출력합니다.
func (l *StandardLogger) Debug(component ComponentType, format string, v ...interface{}) {
	if l.level <= DebugLevel {
		l.logger.Printf("[%s][DEBUG] %s", component, fmt.Sprintf(format, v...))
	}
}

// Info는 정보 레벨 로그를 출력합니다.
func (l *StandardLogger) Info(component ComponentType, format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.logger.Printf("[%s][INFO] %s", component, fmt.Sprintf(format, v...))
	}
}

// Warn은 경고 레벨 로그를 출력합니다.
func (l *StandardLogger) Warn(component ComponentType, format string, v ...interface{}) {
	if l.level <= WarnLevel {
		l.logger.Printf("[%s][WARN] %s", component, fmt.Sprintf(format, v...))
	}
}

// Error는 오류 레벨 로그를 출력합니다.
func (l *StandardLogger) Error(component ComponentType, format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		l.logger.Printf("[%s][ERROR] %s", component, fmt.Sprintf(format, v...))
	}
}

// Timed는 작업 시간을 측정하는 로거 함수를 반환합니다.
func (l *StandardLogger) Timed(component ComponentType, format string, v ...interface{}) func(result string) {
	if l.level > InfoLevel {
		return func(string) {}
	}

	start := time.Now()
	l.Info(component, format, v...)

	return func(result string) {
		elapsed := time.Since(start)
		l.Info(component, "%s: %s (소요 시간: %s)", fmt.Sprintf(format, v...), result, elapsed)
	}
}

// Buffer는 로그 캡처를 위한 버퍼 구조체입니다.
type Buffer struct {
	buf bytes.Buffer
}

// NewBuffer는 새로운 로그 버퍼를 생성합니다.
func NewBuffer() *Buffer {
	return &Buffer{}
}

// Write는 bytes.Buffer Write 메서드를 구현합니다.
func (lb *Buffer) Write(p []byte) (n int, err error) {
	return lb.buf.Write(p)
}

// String은 로그 버퍼의 내용을 문자열로 반환합니다.
func (lb *Buffer) String() string {
	return lb.buf.String()
}

// CaptureLogsToBuffer는 로그 출력을 버퍼로 전환합니다.
// 반환된 함수를 호출하여 원래 출력으로 복원합니다.
func CaptureLogsToBuffer(buffer *Buffer) func() {
	original := log.Writer()
	log.SetOutput(buffer)
	return func() {
		log.SetOutput(original)
	}
}

// ConfigureLogger는 로그 형식과 출력을 설정합니다.
func ConfigureLogger(flags LogFlags, prefix string) {
	var logFlags int

	if flags.Date {
		logFlags |= log.Ldate
	}
	if flags.Time {
		logFlags |= log.Ltime
	}
	if flags.Microseconds {
		logFlags |= log.Lmicroseconds
	}
	if flags.UTC {
		logFlags |= log.LUTC
	}
	if flags.File {
		if flags.LongFile {
			logFlags |= log.Llongfile
		} else {
			logFlags |= log.Lshortfile
		}
	}

	log.SetFlags(logFlags)
	log.SetPrefix(prefix)

	// 기본 로거 업데이트
	defaultLogger = log.Default()
	if stdLogger, ok := sharedLogger.(*StandardLogger); ok {
		stdLogger.logger = defaultLogger
	} else {
		sharedLogger = &StandardLogger{logger: defaultLogger, level: currentLevel}
	}
}

// ConfigureDefaultLogger는 일반적인 로그 형식으로 설정합니다.
func ConfigureDefaultLogger() {
	ConfigureLogger(LogFlags{
		Date: true,
		Time: true,
		File: true,
	}, "[FLOO] ")
}
