package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type LogFormat string

const (
	FormatText LogFormat = "text"
	FormatJSON LogFormat = "json"
)

type Config struct {
	Level  string
	Format LogFormat
	Output io.Writer
}

type Logger struct {
	slog *slog.Logger
}

func New(cfg *Config) *Logger {
	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	level := parseLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.Format == FormatJSON {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return &Logger{slog: slog.New(handler)}
}

func Default() *Logger {
	return New(&Config{
		Level:  "info",
		Format: FormatText,
	})
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{slog: l.slog.With(args...)}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, msg, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, msg, args...)
}

func (l *Logger) LogPhase(jobID string, phase string, args ...any) {
	allArgs := make([]any, 0, len(args)+2)
	allArgs = append(allArgs, "job_id", jobID)
	allArgs = append(allArgs, args...)
	l.slog.Info(phase, allArgs...)
}

func (l *Logger) LogPhaseError(jobID string, phase string, err error, args ...any) {
	allArgs := make([]any, 0, len(args)+4)
	allArgs = append(allArgs, "job_id", jobID)
	allArgs = append(allArgs, "error", err)
	allArgs = append(allArgs, args...)
	l.slog.Error(phase, allArgs...)
}

func (l *Logger) LogDetection(strategy string, severity string, score float64) {
	l.slog.Info("detection",
		"strategy", strategy,
		"severity", severity,
		"score", score,
	)
}

func (l *Logger) LogAnalysis(sourceType string, sourceID string, args ...any) {
	allArgs := make([]any, 0, len(args)+4)
	allArgs = append(allArgs, "source_type", sourceType)
	allArgs = append(allArgs, "source_id", sourceID)
	allArgs = append(allArgs, args...)
	l.slog.Info("analysis", allArgs...)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
