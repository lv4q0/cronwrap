package middleware_test

import (
	"context"
	"os"
	"testing"

	"github.com/example/cronwrap/internal/middleware"
)

func TestEnvMiddlewareInjectsIntoContext(t *testing.T) {
	vars := map[string]string{"FOO": "bar", "BAZ": "qux"}
	var captured map[string]string

	job := func(ctx context.Context) error {
		captured = middleware.EnvFromContext(ctx)
		return nil
	}

	wrapped := middleware.EnvMiddleware(vars, false)(job)
	if err := wrapped(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured["FOO"] != "bar" || captured["BAZ"] != "qux" {
		t.Errorf("expected injected vars, got %v", captured)
	}
}

func TestEnvMiddlewareSetsProcessEnv(t *testing.T) {
	vars := map[string]string{"CRONWRAP_TEST_KEY": "testvalue"}

	var seenDuring string
	job := func(ctx context.Context) error {
		seenDuring = os.Getenv("CRONWRAP_TEST_KEY")
		return nil
	}

	wrapped := middleware.EnvMiddleware(vars, true)(job)
	if err := wrapped(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if seenDuring != "testvalue" {
		t.Errorf("expected process env to be set during job, got %q", seenDuring)
	}
	if after := os.Getenv("CRONWRAP_TEST_KEY"); after != "" {
		t.Errorf("expected process env to be restored after job, got %q", after)
	}
}

func TestEnvMiddlewareRestoresOriginalValue(t *testing.T) {
	os.Setenv("CRONWRAP_RESTORE_KEY", "original")
	t.Cleanup(func() { os.Unsetenv("CRONWRAP_RESTORE_KEY") })

	vars := map[string]string{"CRONWRAP_RESTORE_KEY": "overridden"}
	job := func(ctx context.Context) error { return nil }

	wrapped := middleware.EnvMiddleware(vars, true)(job)
	_ = wrapped(context.Background())

	if got := os.Getenv("CRONWRAP_RESTORE_KEY"); got != "original" {
		t.Errorf("expected original value restored, got %q", got)
	}
}

func TestEnvFromContextEmpty(t *testing.T) {
	got := middleware.EnvFromContext(context.Background())
	if got != nil {
		t.Errorf("expected nil for empty context, got %v", got)
	}
}

func TestFormatEnvNonEmpty(t *testing.T) {
	vars := map[string]string{"key": "val"}
	slice := middleware.FormatEnv(vars)
	if len(slice) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(slice))
	}
	if slice[0] != "KEY=val" {
		t.Errorf("unexpected format: %q", slice[0])
	}
}

func TestEnvMiddlewarePropagatesError(t *testing.T) {
	vars := map[string]string{"X": "1"}
	wantErr := context.DeadlineExceeded

	job := func(ctx context.Context) error { return wantErr }
	wrapped := middleware.EnvMiddleware(vars, false)(job)

	if err := wrapped(context.Background()); err != wantErr {
		t.Errorf("expected %v, got %v", wantErr, err)
	}
}
