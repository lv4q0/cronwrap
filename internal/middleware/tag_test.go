package middleware

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestTagMiddlewareInjectsTagsIntoContext(t *testing.T) {
	tags := Tags{"env": "prod", "team": "platform"}
	mw := TagMiddleware(tags)

	var got Tags
	job := func(ctx context.Context) error {
		got = TagsFromContext(ctx)
		return nil
	}

	if err := mw(job)(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", got["env"])
	}
	if got["team"] != "platform" {
		t.Errorf("expected team=platform, got %q", got["team"])
	}
}

func TestTagMiddlewarePropagatesJobError(t *testing.T) {
	mw := TagMiddleware(Tags{"x": "1"})
	want := errors.New("boom")
	err := mw(func(ctx context.Context) error { return want })(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestTagsFromContextEmpty(t *testing.T) {
	tags := TagsFromContext(context.Background())
	if len(tags) != 0 {
		t.Errorf("expected empty tags, got %v", tags)
	}
}

func TestFormatTagsNonEmpty(t *testing.T) {
	s := FormatTags(Tags{"env": "staging"})
	if !strings.Contains(s, "env=staging") {
		t.Errorf("unexpected format: %q", s)
	}
}

func TestFormatTagsEmpty(t *testing.T) {
	if s := FormatTags(Tags{}); s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

func TestTagMiddlewareChainedWithLog(t *testing.T) {
	tags := Tags{"job": "nightly"}
	var captured Tags

	inner := func(ctx context.Context) error {
		captured = TagsFromContext(ctx)
		return nil
	}

	chain := Chain(
		TagMiddleware(tags),
	)
	if err := Apply(chain, inner)(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured["job"] != "nightly" {
		t.Errorf("tag not propagated through chain")
	}
}
