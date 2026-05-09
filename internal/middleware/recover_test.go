package middleware_test

import (
	"context"
	"strings"
	"testing"

	"github.com/yourorg/cronwrap/internal/middleware"
)

func TestRecoverMiddlewareCatchesPanic(t *testing.T) {
	fn := middleware.Apply(
		func(ctx context.Context) error {
			panic("something went wrong")
		},
		middleware.RecoverMiddleware(),
	)

	err := fn(context.Background())
	if err == nil {
		t.Fatal("expected error from panic recovery, got nil")
	}
	if !strings.Contains(err.Error(), "job panicked") {
		t.Fatalf("unexpected error message: %v", err)
	}
	if !strings.Contains(err.Error(), "something went wrong") {
		t.Fatalf("panic value not included in error: %v", err)
	}
}

func TestRecoverMiddlewarePassesThroughNilError(t *testing.T) {
	fn := middleware.Apply(
		func(ctx context.Context) error { return nil },
		middleware.RecoverMiddleware(),
	)

	if err := fn(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecoverMiddlewarePreservesNonPanicError(t *testing.T) {
	expected := "regular error"
	fn := middleware.Apply(
		func(ctx context.Context) error { return fmt.Errorf("%s", expected) },
		middleware.RecoverMiddleware(),
	)

	err := fn(context.Background())
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected %q, got %v", expected, err)
	}
}
