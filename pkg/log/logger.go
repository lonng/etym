package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"etym/pkg/stack"
	"etym/pkg/tags"
)

type LogLevel byte

const (
	// LevelPanic level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	LevelPanic LogLevel = iota
	// LevelFatal level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	LevelFatal
	// LevelError level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	LevelError
	// LevelWarn level. Non-critical entries that deserve eyes.
	LevelWarn
	// LevelInfo level. General operational entries about what's going on inside the
	// application.
	LevelInfo
	// LevelDebug level. Usually only enabled when debugging. Very verbose logging.
	LevelDebug
)

var (
	textFormatter    = "2006-01-02T15:04:05.000Z07:00"
	consoleFormatter = "15:04:05.000"
)

type Logger struct {
	*log.Logger
	level LogLevel
	dp    int
}

func init() {
	flag := log.LstdFlags | log.Lmicroseconds
	if tags.DEBUG {
		flag |= log.Llongfile
	} else {
		flag |= log.Lshortfile
	}
	std.SetFlags(flag)
}

var std = &Logger{
	Logger: log.New(os.Stderr, "", log.LstdFlags),
	level:  LevelInfo,
	dp:     1,
}

func SetOutput(output io.Writer) {
	std.SetOutput(output)
}

func Std() *Logger { return std }

func ParseLevel(level string) LogLevel {
	level = strings.ToLower(level)
	level = strings.TrimSpace(level)
	switch level {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "panic":
		return LevelPanic
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// 设置日志输出级别
func SetLevel(level string) {
	std.SetLevel(ParseLevel(level))
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

const depth = 2

// Debug logs a message at level Debug on the standard logger.
func (l *Logger) Debug(args ...interface{}) {
	if l.level >= LevelDebug {
		l.Output(depth+l.dp, fmt.Sprint(args...))
	}
}

// Info logs a message at level Info on the standard logger.
func (l *Logger) Info(args ...interface{}) {
	if l.level >= LevelInfo {
		l.Output(depth+l.dp, fmt.Sprint(args...))
	}
}

// Warn logs a message at level Warn on the standard logger.
func (l *Logger) Warn(args ...interface{}) {
	if l.level >= LevelWarn {
		l.Output(depth+l.dp, fmt.Sprint(args...))
	}
}

// Error logs a message at level Error on the standard logger.
func (l *Logger) Error(args ...interface{}) {
	if l.level >= LevelError {
		l.Output(depth+l.dp, fmt.Sprintln(args...)+stack.Backtrace(depth+l.dp))
	}
}

// Panic logs a message at level Panic on the standard logger.
func (l *Logger) Panic(args ...interface{}) {
	if l.level >= LevelPanic {
		l.Output(depth+l.dp, fmt.Sprint(args...))
	}
}

// Fatal logs a message at level Fatal on the standard logger.
func (l *Logger) Fatal(args ...interface{}) {
	if l.level >= LevelFatal {
		l.Output(depth+l.dp, fmt.Sprint(args...))
	}
}

// Debugf logs a message at level Debug on the standard logger.
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.level >= LevelDebug {
		l.Output(depth+l.dp, fmt.Sprintf(format, args...))
	}
}

// Infof logs a message at level Info on the standard logger.
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.level >= LevelInfo {
		l.Output(depth+l.dp, fmt.Sprintf(format, args...))
	}
}

// Warnf logs a message at level Warn on the standard logger.
func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.level >= LevelWarn {
		l.Output(depth+l.dp, fmt.Sprintf(format, args...))
	}
}

// Errorf logs a message at level Error on the standard logger.
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.level >= LevelError {
		l.Output(depth+l.dp, fmt.Sprintf(format+"\n", args...)+stack.Backtrace(depth+l.dp))
	}
}

// Panicf logs a message at level Panic on the standard logger.
func (l *Logger) Panicf(format string, args ...interface{}) {
	if l.level >= LevelPanic {
		l.Output(depth+l.dp, fmt.Sprintf(format, args...))
	}
}

// Fatalf logs a message at level Fatal on the standard logger.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	if l.level >= LevelFatal {
		l.Output(depth+l.dp, fmt.Sprintf(format, args...))
	}
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	std.Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	std.Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	std.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	std.Warn(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	std.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	std.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	std.Fatal(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(format string, args ...interface{}) {
	std.Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	std.Warnf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	std.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	std.Fatalf(format, args...)
}
