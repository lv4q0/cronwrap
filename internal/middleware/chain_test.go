package middleware_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourorg/cronwrap/internal/middleware"
)

func TestChainCallsMiddlewaresInOrder(t *testing.T) {
	var order []int

	make := func(id int) middleware.Middleware {
		return func(next middleware.RunFunc) middleware.RunFunc {
			return func(ctx context.Context) error {
				order = append(order, id)
				return next(ctx)
			}
		}
	}

	fn := middleware.Apply(
		func(ctx context.Context) error { return nil },
		make(1), make(2), make(3),
	)

	if err := fn(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Fatalf("expected order [1 2 3], got %v", order)
	}
}

func TestChainPropagatesError(t *testing.T) {
	sentinel := errors.New("boom")

	fn := middleware.Apply(
		func(ctx context.Context) error { return sentinel },
	)

	if err := fn(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestChainEmptyIsIdentity(t *testing.T) {
	called := false
	fn := middleware.Apply(func(ctx context.Context) error {
		called = true
		return nil
	})

	_ = fn(context.Background())
	if !called {
		t.Fatal("expected inner function to be called")
	}
}

func TestChainShortCircuitsOnError(t *testing.T) {
	sentinel := errors.New("early")
	innerCalled := false

	failMiddleware := func(next middleware.RunFunc) middleware.RunFunc {
		return func(ctx context.Context) error {
			return sentinel
		}
	}

	fn := middleware.Apply(
		func(ctx context.Context) error { innerCalled = true; return nil },
		failMiddleware,
	)

	if err := fn(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}
	if innerCalled {
		t.Fatal("inner function should not have been called")
	}
}
