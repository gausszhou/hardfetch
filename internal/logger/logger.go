package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

var (
	logger   *slog.Logger
	loggerMu sync.RWMutex
	isDebug  bool
)

func Init(debug bool) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	isDebug = debug
	if debug {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
}

func Get() *slog.Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
	return logger
}

func IsDebug() bool {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return isDebug
}

func Debug(ctx context.Context, msg string, args ...any) {
	if IsDebug() {
		Get().DebugContext(ctx, msg, args...)
	}
}

func Info(ctx context.Context, msg string, args ...any) {
	Get().InfoContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	Get().ErrorContext(ctx, msg, args...)
}

func DebugFunc(name string, fn func() error) error {
	if !IsDebug() {
		return fn()
	}
	start := time.Now()
	Get().Debug("executing", "function", name)
	err := fn()
	duration := time.Since(start)
	if err != nil {
		Get().Debug("function completed with error", "function", name, "duration_ms", duration.Milliseconds(), "error", err)
	} else {
		Get().Debug("function completed", "function", name, "duration_ms", duration.Milliseconds())
	}
	return err
}

func DebugFuncResult[T any](name string, fn func() (T, error)) (T, error) {
	if !IsDebug() {
		return fn()
	}
	start := time.Now()
	Get().Debug("executing", "function", name)
	result, err := fn()
	duration := time.Since(start)
	if err != nil {
		Get().Debug("function completed with error", "function", name, "duration_ms", duration.Milliseconds(), "error", err)
	} else {
		Get().Debug("function completed", "function", name, "duration_ms", duration.Milliseconds())
	}
	return result, err
}

type Timer struct {
	name      string
	startTime time.Time
	logger    *slog.Logger
}

func StartTimer(name string) *Timer {
	t := &Timer{
		name:      name,
		startTime: time.Now(),
		logger:    Get(),
	}
	if IsDebug() {
		t.logger.Debug("timer started", "name", name)
	}
	return t
}

func (t *Timer) Stop() {
	if IsDebug() {
		duration := time.Since(t.startTime)
		t.logger.Debug("timer stopped", "name", t.name, "duration_ms", duration.Milliseconds())
	}
}

func (t *Timer) Checkpoint(name string) {
	if IsDebug() {
		duration := time.Since(t.startTime)
		t.logger.Debug("timer checkpoint", "name", t.name, "checkpoint", name, "elapsed_ms", duration.Milliseconds())
	}
}

func FormatDuration(d time.Duration) string {
	return fmt.Sprintf("%.2fms", d.Seconds()*1000)
}
