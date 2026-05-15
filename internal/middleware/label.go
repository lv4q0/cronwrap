package middleware

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type labelKey struct{}

// LabelMiddleware injects a static map of key/value labels into the job context.
// Labels differ from tags in that they are arbitrary string pairs rather than
// free-form string tokens, making them suitable for structured metadata such as
// team ownership, environment tier, or cost-centre codes.
func LabelMiddleware(labels map[string]string) func(next func(context.Context) error) func(context.Context) error {
	return func(next func(context.Context) error) func(context.Context) error {
		return func(ctx context.Context) error {
			merged := make(map[string]string)
			if existing := LabelsFromContext(ctx); existing != nil {
				for k, v := range existing {
					merged[k] = v
				}
			}
			for k, v := range labels {
				merged[k] = v
			}
			ctx = context.WithValue(ctx, labelKey{}, merged)
			return next(ctx)
		}
	}
}

// LabelsFromContext retrieves the label map stored in ctx by LabelMiddleware.
// Returns nil when no labels are present.
func LabelsFromContext(ctx context.Context) map[string]string {
	v, _ := ctx.Value(labelKey{}).(map[string]string)
	return v
}

// FormatLabels returns a deterministic, human-readable representation of the
// label map, e.g. "env=prod team=platform". Returns an empty string when the
// map is nil or empty.
func FormatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}
	return strings.Join(parts, " ")
}
