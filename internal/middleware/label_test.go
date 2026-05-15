package middleware_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cronwrap/cronwrap/internal/middleware"
)

func TestLabelMiddlewareInjectsLabelsIntoContext(t *testing.T) {
	labels := map[string]string{"env": "prod", "team": "platform"}
	var captured map[string]string

	job := func(ctx context.Context) error {
		captured = middleware.LabelsFromContext(ctx)
		return nil
	}

	wrapped := middleware.LabelMiddleware(labels)(job)
	if err := wrapped(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured["env"] != "prod" || captured["team"] != "platform" {
		t.Errorf("expected labels injected, got %v", captured)
	}
}

func TestLabelMiddlewarePropagatesJobError(t *testing.T) {
	sentinel := errors.New("boom")
	job := func(ctx context.Context) error { return sentinel }

	wrapped := middleware.LabelMiddleware(map[string]string{"k": "v"})(job)
	if err := wrapped(context.Background()); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestLabelMiddlewareMergesWithExistingLabels(t *testing.T) {
	outer := middleware.LabelMiddleware(map[string]string{"env": "prod"})
	inner := middleware.LabelMiddleware(map[string]string{"team": "sre"})

	var captured map[string]string
	job := func(ctx context.Context) error {
		captured = middleware.LabelsFromContext(ctx)
		return nil
	}

	wrapped := outer(inner(job))
	if err := wrapped(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured["env"] != "prod" || captured["team"] != "sre" {
		t.Errorf("expected merged labels, got %v", captured)
	}
}

func TestLabelsFromContextEmpty(t *testing.T) {
	if got := middleware.LabelsFromContext(context.Background()); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestFormatLabelsNonEmpty(t *testing.T) {
	labels := map[string]string{"env": "staging", "region": "eu-west"}
	got := middleware.FormatLabels(labels)
	want := "env=staging region=eu-west"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestFormatLabelsEmpty(t *testing.T) {
	if got := middleware.FormatLabels(nil); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
