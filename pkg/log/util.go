package log

import (
	"bytes"
	"log"
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

var (
	// 현재 로그 레벨, 기본값은 InfoLevel
	currentLevel = InfoLevel
	// 기본 로거
	defaultLogger = log.Default()
)

// SetLogLevel은 현재 로그 레벨을 설정합니다.
func SetLogLevel(level LogLevel) {
	currentLevel = level
}

// GetLogLevel은 현재 로그 레벨을 반환합니다.
func GetLogLevel() LogLevel {
	return currentLevel
}

// IsDebugEnabled는 디버그 로그가 활성화되었는지 확인합니다.
func IsDebugEnabled() bool {
	return currentLevel <= DebugLevel
}

// GetLogger는 기본 로거를 반환합니다.
func GetLogger() *log.Logger {
	return defaultLogger
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

	// 기본 로거 생성
	defaultLogger = log.Default()
}

// ConfigureDefaultLogger는 일반적인 로그 형식으로 설정합니다.
func ConfigureDefaultLogger() {
	ConfigureLogger(LogFlags{
		Date: true,
		Time: true,
		File: true,
	}, "[FLOO] ")
}
