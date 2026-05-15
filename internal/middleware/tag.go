package middleware

import (
	"context"
	"fmt"
)

// TagKey is the context key used to store job tags.
type TagKey struct{}

// Tags holds a set of key/value labels attached to a job execution.
type Tags map[string]string

// TagMiddleware injects a set of static tags into the context so that
// downstream middleware and handlers can read them for logging, metrics, etc.
func TagMiddleware(tags Tags) func(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(next func(ctx context.Context) error) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			ctx = context.WithValue(ctx, TagKey{}, tags)
			return next(ctx)
		}
	}
}

// TagsFromContext retrieves the Tags stored in ctx.
// Returns an empty Tags map when none are present.
func TagsFromContext(ctx context.Context) Tags {
	v := ctx.Value(TagKey{})
	if v == nil {
		return Tags{}
	}
	t, ok := v.(Tags)
	if !ok {
		return Tags{}
	}
	return t
}

// FormatTags returns a human-readable representation of the tag map,
// e.g. "env=prod team=platform".
func FormatTags(t Tags) string {
	if len(t) == 0 {
		return ""
	}
	out := ""
	for k, v := range t {
		if out != "" {
			out += " "
		}
		out += fmt.Sprintf("%s=%s", k, v)
	}
	return out
}
