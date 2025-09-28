package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

// Logger 封装zerolog日志器
type Logger struct {
	logger    zerolog.Logger
	Module    string
	Submodule string
	Ctx       context.Context
}

type Option func(l *Logger)

func WithModule(m string) Option {
	return func(l *Logger) {
		l.Module = m
	}
}

func WithSubmodule(s string) Option {
	return func(l *Logger) {
		l.Submodule = s
	}
}

func WithCtx(ctx context.Context) Option {
	return func(l *Logger) {
		l.Ctx = ctx
	}
}

func WithLogger(logger zerolog.Logger) Option {
	return func(l *Logger) {
		l.logger = logger
	}
}

// NewLogger 创建新的日志器实例
func NewLogger() *Logger {
	// 设置全局日志级别
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// 创建带颜色的控制台输出
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	}

	// 创建logger实例
	logger := zerolog.New(output).With().Timestamp().Logger()

	return &Logger{logger: logger}
}

// SetDebugLevel 设置为调试级别
func (l *Logger) SetDebugLevel() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

// SetQuietMode 设置为静默模式（只显示错误）
func (l *Logger) SetQuietMode() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

func (l *Logger) Info() *zerolog.Event {
	ev := l.logger.Info()

	if l.Module != "" {
		ev = ev.Str("Module", l.Module)
	}
	if l.Submodule != "" {
		ev = ev.Str("SubModule", l.Submodule)
	}
	return ev
}

func (l *Logger) Debug() *zerolog.Event {
	ev := l.logger.Debug()
	if l.Module != "" {
		ev = ev.Str("Module", l.Module)
	}
	if l.Submodule != "" {
		ev = ev.Str("SubModule", l.Submodule)
	}
	return ev
}

func (l *Logger) Trace() *zerolog.Event {
	ev := l.logger.Trace()
	if l.Module != "" {
		ev = ev.Str("Module", l.Module)
	}
	if l.Submodule != "" {
		ev = ev.Str("SubModule", l.Submodule)
	}
	return ev
}

func (l *Logger) Err(err error) *zerolog.Event {
	ev := l.logger.Err(err)
	if l.Module != "" {
		ev = ev.Str("Module", l.Module)
	}
	if l.Submodule != "" {
		ev = ev.Str("SubModule", l.Submodule)
	}
	return ev
}

func (l *Logger) Warn() *zerolog.Event {
	ev := l.logger.Warn()
	if l.Module != "" {
		ev = ev.Str("Module", l.Module)
	}
	if l.Submodule != "" {
		ev = ev.Str("SubModule", l.Submodule)
	}
	return ev
}

func (l *Logger) Error() *zerolog.Event {
	ev := l.logger.Error()
	if l.Module != "" {
		ev = ev.Str("Module", l.Module)
	}
	if l.Submodule != "" {
		ev = ev.Str("SubModule", l.Submodule)
	}
	return ev
}

// UserInfo 输出用户友好的信息消息（直接输出到stdout，不带时间戳）
func (l *Logger) UserInfo(msg string) {
	fmt.Println(msg)
}

// UserError 输出用户友好的错误消息（直接输出到stderr，不带时间戳）
func (l *Logger) UserError(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
}

// UserPrompt 输出用户提示信息（不换行，用于交互式输入）
func (l *Logger) UserPrompt(msg string) {
	fmt.Print(msg)
}

// UserSuccess 输出成功消息
func (l *Logger) UserSuccess(msg string) {
	fmt.Println(msg)
}

// UserWarning 输出警告消息
func (l *Logger) UserWarning(msg string) {
	fmt.Println(msg)
}
