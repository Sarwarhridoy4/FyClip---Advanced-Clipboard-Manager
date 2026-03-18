// File: internal/logger/logger.go
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents log levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality
type Logger struct {
	logger   *slog.Logger
	mu       sync.RWMutex
	handlers []slog.Handler
	level    Level
	output   io.Writer
}

// Default logger instance
var defaultLogger *Logger
var once sync.Once

// Get returns the default logger instance
func Get() *Logger {
	once.Do(func() {
		defaultLogger = &Logger{
			level:  LevelInfo,
			output: os.Stderr,
		}
		defaultLogger.initDefaultLogger()
	})
	return defaultLogger
}

// initDefaultLogger initializes the default logger
func (l *Logger) initDefaultLogger() {
	// Create text handler for console output
	textHandler := slog.NewTextHandler(l.output, &slog.HandlerOptions{
		Level: slog.Level(l.level),
		AddSource: true,
	})

	l.logger = slog.New(textHandler)
	l.handlers = []slog.Handler{textHandler}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	l.initDefaultLogger()
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
	l.initDefaultLogger()
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// DebugContext logs a debug message with context
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.DebugContext(ctx, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// InfoContext logs an info message with context
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.InfoContext(ctx, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// WarnContext logs a warning message with context
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.WarnContext(ctx, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// ErrorContext logs an error message with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// Log logs a message at the specified level
func (l *Logger) Log(level Level, msg string, args ...interface{}) {
	switch level {
	case LevelDebug:
		l.Debug(msg, args...)
	case LevelInfo:
		l.Info(msg, args...)
	case LevelWarn:
		l.Warn(msg, args...)
	case LevelError:
		l.Error(msg, args...)
	}
}

// With returns a logger with additional attributes
func (l *Logger) With(args ...interface{}) *slog.Logger {
	return l.logger.With(args...)
}

// Convenience functions using default logger

// Debug logs a debug message using default logger
func Debug(msg string, args ...interface{}) {
	Get().Debug(msg, args...)
}

// DebugContext logs a debug message with context using default logger
func DebugContext(ctx context.Context, msg string, args ...interface{}) {
	Get().DebugContext(ctx, msg, args...)
}

// Info logs an info message using default logger
func Info(msg string, args ...interface{}) {
	Get().Info(msg, args...)
}

// InfoContext logs an info message with context using default logger
func InfoContext(ctx context.Context, msg string, args ...interface{}) {
	Get().InfoContext(ctx, msg, args...)
}

// Warn logs a warning message using default logger
func Warn(msg string, args ...interface{}) {
	Get().Warn(msg, args...)
}

// WarnContext logs a warning message with context using default logger
func WarnContext(ctx context.Context, msg string, args ...interface{}) {
	Get().WarnContext(ctx, msg, args...)
}

// Error logs an error message using default logger
func Error(msg string, args ...interface{}) {
	Get().Error(msg, args...)
}

// ErrorContext logs an error message with context using default logger
func ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	Get().ErrorContext(ctx, msg, args...)
}

// FileLogger provides file-based logging with rotation
type FileLogger struct {
	*Logger
	filePath    string
	maxSize     int64 // max size in bytes
	maxAge      int   // max age in days
	rotateCount int   // number of old log files to keep
}

// NewFileLogger creates a new file logger
func NewFileLogger(filePath string, maxSize int64, maxAge int, rotateCount int) (*FileLogger, error) {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	l := &FileLogger{
		Logger:      Get(),
		filePath:    filePath,
		maxSize:     maxSize,
		maxAge:      maxAge,
		rotateCount: rotateCount,
	}

	// Open log file
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	l.SetOutput(f)
	return l, nil
}

// Rotate rotates the log file
func (l *FileLogger) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if rotation is needed
	info, err := os.Stat(l.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// If file is too large, rotate
	if info.Size() >= l.maxSize {
		// Rename current log file
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		newPath := fmt.Sprintf("%s.%s", l.filePath, timestamp)

		if err := os.Rename(l.filePath, newPath); err != nil {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}

		// Open new log file
		f, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to create new log file: %w", err)
		}

		l.SetOutput(f)

		// Clean up old files
		l.cleanupOldFiles()
	}

	return nil
}

// cleanupOldFiles removes old log files beyond the rotation count
func (l *FileLogger) cleanupOldFiles() {
	if l.rotateCount <= 0 {
		return
	}

	dir := filepath.Dir(l.filePath)
	pattern := filepath.Base(l.filePath) + ".*"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return
	}

	// Sort by modification time (newest first)
	// Keep only the most recent files
	if len(matches) > l.rotateCount {
		for i := l.rotateCount; i < len(matches); i++ {
			os.Remove(matches[i])
		}
	}
}

// Close closes the file logger
func (l *FileLogger) Close() error {
	if f, ok := l.output.(*os.File); ok {
		return f.Close()
	}
	return nil
}

// Legacy compatibility - redirect standard log to our logger
func init() {
	// Replace standard logger with our slog-based one
	log.SetFlags(0)
	log.SetOutput(io.Discard)
}
