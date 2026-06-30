package logger

import (
	"context"
	"errors"
	"testing"
)

func TestLogError_NilErrorDoesNotPanic(t *testing.T) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LogError panicked with nil error: %v", r)
		}
	}()

	LogError(context.Background(), nil, "test nil error")
}

func TestLogError_WithErrorDoesNotPanic(t *testing.T) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LogError panicked with error value: %v", r)
		}
	}()

	LogError(context.Background(), errors.New("boom"), "test error")
}
