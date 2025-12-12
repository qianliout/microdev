package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger    zerolog.Logger
	out       io.Writer
	Module    string
	Submodule string
}

type Option func(l *Logger)

func NewLogger(opts ...Option) *Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	console := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05"}

	l := &Logger{out: console}

	for _, opt := range opts {
		opt(l)
	}

	c := zerolog.New(l.out).With().Timestamp()
	if l.Module != "" {
		c.Str("Module", l.Module)
	}
	if l.Submodule != "" {
		c.Str("Submodule", l.Submodule)
	}
	l.logger = c.Logger()

	return l
}

func (l *Logger) SetDebugLevel() { zerolog.SetGlobalLevel(zerolog.DebugLevel) }
func (l *Logger) SetQuietMode()  { zerolog.SetGlobalLevel(zerolog.ErrorLevel) }

func (l *Logger) Info() *zerolog.Event         { return l.logger.Info() }
func (l *Logger) Debug() *zerolog.Event        { return l.logger.Debug() }
func (l *Logger) Trace() *zerolog.Event        { return l.logger.Trace() }
func (l *Logger) Warn() *zerolog.Event         { return l.logger.Warn() }
func (l *Logger) Error() *zerolog.Event        { return l.logger.Error() }
func (l *Logger) Err(err error) *zerolog.Event { return l.logger.Err(err) }
func (l *Logger) Writer() io.Writer            { return l.out }
