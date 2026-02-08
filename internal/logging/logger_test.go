package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		logger := New(nil)
		if logger == nil {
			t.Fatal("New(nil) returned nil")
		}
	})

	t.Run("json format", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New(&Config{
			Level:  "info",
			Format: FormatJSON,
			Output: &buf,
		})
		logger.Info("test message", "key", "value")
		output := buf.String()
		if !strings.Contains(output, `"msg":"test message"`) {
			t.Errorf("expected JSON output, got: %s", output)
		}
		if !strings.Contains(output, `"key":"value"`) {
			t.Errorf("expected key=value in output, got: %s", output)
		}
	})

	t.Run("text format", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New(&Config{
			Level:  "info",
			Format: FormatText,
			Output: &buf,
		})
		logger.Info("test message", "key", "value")
		output := buf.String()
		if !strings.Contains(output, "test message") {
			t.Errorf("expected 'test message' in output, got: %s", output)
		}
	})

	t.Run("debug level filters info when set to error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New(&Config{
			Level:  "error",
			Format: FormatText,
			Output: &buf,
		})
		logger.Info("should not appear")
		if buf.Len() > 0 {
			t.Errorf("expected empty output for info at error level, got: %s", buf.String())
		}
	})
}

func TestDefault(t *testing.T) {
	logger := Default()
	if logger == nil {
		t.Fatal("Default() returned nil")
	}
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&Config{
		Level:  "info",
		Format: FormatJSON,
		Output: &buf,
	})

	child := logger.With("component", "webhook")
	child.Info("test")
	output := buf.String()
	if !strings.Contains(output, `"component":"webhook"`) {
		t.Errorf("expected component=webhook in output, got: %s", output)
	}
}

func TestLogger_LogPhase(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&Config{
		Level:  "info",
		Format: FormatJSON,
		Output: &buf,
	})

	logger.LogPhase("job-123", "cloning", "repo", "https://github.com/example/repo")
	output := buf.String()
	if !strings.Contains(output, `"job_id":"job-123"`) {
		t.Errorf("expected job_id in output, got: %s", output)
	}
	if !strings.Contains(output, `"msg":"cloning"`) {
		t.Errorf("expected msg=cloning in output, got: %s", output)
	}
	if !strings.Contains(output, `"repo":"https://github.com/example/repo"`) {
		t.Errorf("expected repo URL in output, got: %s", output)
	}
}

func TestLogger_LogPhaseError(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&Config{
		Level:  "info",
		Format: FormatJSON,
		Output: &buf,
	})

	logger.LogPhaseError("job-456", "clone failed", nil, "repo", "test-repo")
	output := buf.String()
	if !strings.Contains(output, `"level":"ERROR"`) {
		t.Errorf("expected ERROR level in output, got: %s", output)
	}
	if !strings.Contains(output, `"job_id":"job-456"`) {
		t.Errorf("expected job_id in output, got: %s", output)
	}
}

func TestLogger_LogDetection(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&Config{
		Level:  "info",
		Format: FormatJSON,
		Output: &buf,
	})

	logger.LogDetection("velocity-analysis", "high", 0.85)
	output := buf.String()
	if !strings.Contains(output, `"strategy":"velocity-analysis"`) {
		t.Errorf("expected strategy in output, got: %s", output)
	}
	if !strings.Contains(output, `"severity":"high"`) {
		t.Errorf("expected severity in output, got: %s", output)
	}
}

func TestLogger_LogAnalysis(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&Config{
		Level:  "info",
		Format: FormatJSON,
		Output: &buf,
	})

	logger.LogAnalysis("git", "https://github.com/example/repo", "commits", 42)
	output := buf.String()
	if !strings.Contains(output, `"source_type":"git"`) {
		t.Errorf("expected source_type in output, got: %s", output)
	}
	if !strings.Contains(output, `"source_id":"https://github.com/example/repo"`) {
		t.Errorf("expected source_id in output, got: %s", output)
	}
}

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		logFn    func(l *Logger)
		expect   bool
	}{
		{"debug at debug level", "debug", func(l *Logger) { l.Debug("msg") }, true},
		{"info at debug level", "debug", func(l *Logger) { l.Info("msg") }, true},
		{"warn at debug level", "debug", func(l *Logger) { l.Warn("msg") }, true},
		{"error at debug level", "debug", func(l *Logger) { l.Error("msg") }, true},
		{"debug at info level", "info", func(l *Logger) { l.Debug("msg") }, false},
		{"info at info level", "info", func(l *Logger) { l.Info("msg") }, true},
		{"info at warn level", "warn", func(l *Logger) { l.Info("msg") }, false},
		{"warn at warn level", "warn", func(l *Logger) { l.Warn("msg") }, true},
		{"warn at error level", "error", func(l *Logger) { l.Warn("msg") }, false},
		{"error at error level", "error", func(l *Logger) { l.Error("msg") }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&Config{
				Level:  tt.logLevel,
				Format: FormatText,
				Output: &buf,
			})
			tt.logFn(logger)
			hasOutput := buf.Len() > 0
			if hasOutput != tt.expect {
				t.Errorf("level=%s: expected output=%v, got output=%v (%s)", tt.logLevel, tt.expect, hasOutput, buf.String())
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"debug", "DEBUG"},
		{"info", "INFO"},
		{"warn", "WARN"},
		{"warning", "WARN"},
		{"error", "ERROR"},
		{"", "INFO"},
		{"unknown", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got.String() != tt.want {
				t.Errorf("parseLevel(%q) = %s, want %s", tt.input, got.String(), tt.want)
			}
		})
	}
}
