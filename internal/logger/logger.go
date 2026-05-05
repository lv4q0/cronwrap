package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

// Logger writes structured log lines for cron job execution.
type Logger struct {
	out    io.Writer
	level  Level
	jobID  string
}

// New creates a Logger writing to out at the given minimum level.
func New(out io.Writer, level Level, jobID string) *Logger {
	if out == nil {
		out = os.Stderr
	}
	return &Logger{out: out, level: level, jobID: jobID}
}

// NewDefault returns a Logger writing to stderr at INFO level.
func NewDefault(jobID string) *Logger {
	return New(os.Stderr, LevelInfo, jobID)
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	if level < l.level {
		return
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	line := fmt.Sprintf("ts=%s level=%s job=%s msg=%q", ts, levelNames[level], l.jobID, msg)
	for k, v := range fields {
		line += fmt.Sprintf(" %s=%v", k, v)
	}
	fmt.Fprintln(l.out, line)
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string, fields map[string]any) { l.log(LevelDebug, msg, fields) }

// Info logs an info-level message.
func (l *Logger) Info(msg string, fields map[string]any) { l.log(LevelInfo, msg, fields) }

// Warn logs a warn-level message.
func (l *Logger) Warn(msg string, fields map[string]any) { l.log(LevelWarn, msg, fields) }

// Error logs an error-level message.
func (l *Logger) Error(msg string, fields map[string]any) { l.log(LevelError, msg, fields) }
