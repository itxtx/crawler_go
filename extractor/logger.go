package extractor

import (
	"fmt"
	"log"
	"os"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelSilent LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// Logger provides structured logging for extraction operations
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger with the specified level
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

// DefaultLogger is the default logger instance
var DefaultLogger = NewLogger(LogLevelError)

// SetLogLevel sets the logging level for the default logger
func SetLogLevel(level LogLevel) {
	DefaultLogger.level = level
}

// logMessage logs a message at the specified level
func (l *Logger) logMessage(level LogLevel, prefix, format string, args ...interface{}) {
	if l.level >= level {
		message := fmt.Sprintf(format, args...)
		l.logger.Printf("[%s] %s", prefix, message)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.logMessage(LogLevelError, "ERROR", format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.logMessage(LogLevelWarn, "WARN", format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.logMessage(LogLevelInfo, "INFO", format, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.logMessage(LogLevelDebug, "DEBUG", format, args...)
}

// Global logging functions using the default logger
func LogError(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}

func LogWarn(format string, args ...interface{}) {
	DefaultLogger.Warn(format, args...)
}

func LogInfo(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

func LogDebug(format string, args ...interface{}) {
	DefaultLogger.Debug(format, args...)
}

// LogExtractionError logs an ExtractionError with appropriate context
func LogExtractionError(err *ExtractionError, context string) {
	if context != "" {
		LogError("Extraction error in %s: %v", context, err)
	} else {
		LogError("Extraction error: %v", err)
	}
}

// LogExtractionWarn logs an extraction warning
func LogExtractionWarn(message string, args ...interface{}) {
	LogWarn("Extraction warning: "+message, args...)
}

// LogExtractionInfo logs extraction info
func LogExtractionInfo(message string, args ...interface{}) {
	LogInfo("Extraction: "+message, args...)
}

// LogExtractionDebug logs extraction debug info
func LogExtractionDebug(message string, args ...interface{}) {
	LogDebug("Extraction: "+message, args...)
}
