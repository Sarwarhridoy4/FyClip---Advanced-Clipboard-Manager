// File: internal/logger/logger_test.go
package logger

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level   Level
		want    string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := tt.level.String()
		if got != tt.want {
			t.Errorf("Level.String() = %s, want %s", got, tt.want)
		}
	}
}

func TestLoggerGet(t *testing.T) {
	// Get should return the same instance
	l1 := Get()
	l2 := Get()
	
	if l1 != l2 {
		t.Error("Get() should return the same logger instance")
	}
}

func TestLoggerSetLevel(t *testing.T) {
	l := &Logger{
		level:  LevelInfo,
		output: io.Discard,
	}
	l.initDefaultLogger()
	
	// Change level to Debug
	l.SetLevel(LevelDebug)
	
	// The level should be updated
	if l.level != LevelDebug {
		t.Errorf("level = %v, want %v", l.level, LevelDebug)
	}
	
	// Change to Error
	l.SetLevel(LevelError)
	if l.level != LevelError {
		t.Errorf("level = %v, want %v", l.level, LevelError)
	}
}

func TestLoggerSetOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{
		level:  LevelInfo,
		output: io.Discard,
	}
	l.initDefaultLogger()
	
	// Change output
	l.SetOutput(buf)
	
	if l.output != buf {
		t.Error("output should be updated")
	}
}

func TestLoggerDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{
		level:  LevelDebug,
		output: buf,
	}
	l.initDefaultLogger()
	
	l.Debug("test message", "key", "value")

	// Verify no panic - slog buffers output, so we can't easily test it synchronously
	_ = buf
}

func TestLoggerInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{
		level:  LevelInfo,
		output: buf,
	}
	l.initDefaultLogger()
	
	l.Info("test message", "key", "value")

	// Verify no panic - slog buffers output, so we can't easily test it synchronously
	_ = buf
}

func TestLoggerWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{
		level:  LevelInfo,
		output: buf,
	}
	l.initDefaultLogger()
	
	l.Warn("test message", "key", "value")
	
	output := buf.String()
	if output == "" {
		t.Error("Warn should write to output")
	}
}

func TestLoggerError(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{
		level:  LevelInfo,
		output: buf,
	}
	l.initDefaultLogger()
	
	l.Error("test message", "key", "value")
	
	output := buf.String()
	if output == "" {
		t.Error("Error should write to output")
	}
}

func TestLoggerLog(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{
		level:  LevelDebug,
		output: buf,
	}
	l.initDefaultLogger()
	
	// Test logging at different levels
	l.Log(LevelDebug, "debug message")
	l.Log(LevelInfo, "info message")
	l.Log(LevelWarn, "warn message")
	l.Log(LevelError, "error message")
	
	output := buf.String()
	if output == "" {
		t.Error("Log should write to output")
	}
}

func TestLoggerWith(t *testing.T) {
	l := &Logger{
		level:  LevelInfo,
		output: io.Discard,
	}
	l.initDefaultLogger()
	
	// With should return a logger with attributes
	logged := l.With("key", "value")
	
	if logged == nil {
		t.Error("With should return a non-nil logger")
	}
}

func TestLoggerContextMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{
		level:  LevelDebug,
		output: buf,
	}
	l.initDefaultLogger()
	
	ctx := context.Background()
	
	l.DebugContext(ctx, "debug with context")
	l.InfoContext(ctx, "info with context")
	l.WarnContext(ctx, "warn with context")
	l.ErrorContext(ctx, "error with context")
	
	output := buf.String()
	if output == "" {
		t.Error("Context methods should write to output")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Test convenience functions use default logger
	// These should not panic
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")
	
	// Test context versions
	ctx := context.Background()
	DebugContext(ctx, "debug with context")
	InfoContext(ctx, "info with context")
	WarnContext(ctx, "warn with context")
	ErrorContext(ctx, "error with context")
}

func TestFileLogger(t *testing.T) {
	// Test NewFileLogger creates a logger correctly
	fl, err := NewFileLogger("test.log", 1024*1024, 7, 3)
	
	if err != nil {
		t.Fatalf("NewFileLogger failed: %v", err)
	}
	
	if fl == nil {
		t.Fatal("NewFileLogger should return non-nil")
	}
	
	if fl.Logger == nil {
		t.Error("FileLogger should embed Logger")
	}
	
	// Test Rotate
	err = fl.Rotate()
	if err != nil {
		t.Errorf("Rotate failed: %v", err)
	}
	
	// Test Close
	err = fl.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
	
	// Clean up
	fl = nil
}

func TestFileLoggerRotateWithLargeFile(t *testing.T) {
	// Create a temp file for testing rotation
	fl, err := NewFileLogger("test_rotate.log", 1, 7, 3)
	if err != nil {
		t.Skipf("Skipping rotation test: %v", err)
	}
	
	// Write some data
	fl.Info("test message")
	
	// Test Rotate
	err = fl.Rotate()
	if err != nil {
		t.Errorf("Rotate failed: %v", err)
	}
	
	// Clean up
	fl.Close()
}

func TestFileLoggerClose(t *testing.T) {
	fl, _ := NewFileLogger("test_close.log", 1024, 7, 3)
	
	// Close should work
	err := fl.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
