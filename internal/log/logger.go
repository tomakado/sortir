package log

import (
	"log/slog"
	"os"
)

type Logger struct {
	logger *slog.Logger
	level  Level
}

func NewLogger(level Level) *Logger {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level.toSlogLevel()})

	return &Logger{
		logger: slog.New(h),
		level:  level,
	}
}

func (l *Logger) With(args ...any) *Logger {
	newLogger := l.logger.With(args...)
	return &Logger{
		logger: newLogger,
		level:  l.level,
	}
}

func (l *Logger) Important(msg string, args ...any) {
	if l == nil {
		return
	}
	l.Write(Important, msg, args...)
}

func (l *Logger) Verbose(msg string, args ...any) {
	if l == nil {
		return
	}
	l.Write(Verbose, msg, args...)
}

func (l *Logger) Write(level Level, msg string, args ...any) {
	if l == nil {
		return
	}
	switch level {
	case Important:
		l.logger.Error(msg, args...)
	case Verbose:
		l.logger.Debug(msg, args...)
	}
}

type Level uint8

const (
	Important Level = iota
	Verbose
)

func (l Level) toSlogLevel() slog.Level {
	switch l {
	case Important:
		return slog.LevelError
	case Verbose:
		return slog.LevelDebug
	default:
		return slog.LevelDebug
	}
}
