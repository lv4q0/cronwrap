package middleware

import (
	"context"

	"github.com/cronwrap/cronwrap/internal/logger"
)

// TagLoggingMiddleware reads Tags from the context (injected by TagMiddleware)
// and emits them as structured fields at the start and end of each job run.
// It is intended to be placed after TagMiddleware in the chain.
func TagLoggingMiddleware(log *logger.Logger) func(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(next func(ctx context.Context) error) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			tags := TagsFromContext(ctx)
			fields := tagsToFields(tags)

			log.Info("job starting with tags", fields...)
			err := next(ctx)
			if err != nil {
				log.Error("job finished with error", append(fields, "error", err.Error())...)
			} else {
				log.Info("job finished successfully", fields...)
			}
			return err
		}
	}
}

// tagsToFields converts a Tags map into a flat key/value slice suitable for
// structured logger calls.
func tagsToFields(tags Tags) []interface{} {
	if len(tags) == 0 {
		return nil
	}
	out := make([]interface{}, 0, len(tags)*2)
	for k, v := range tags {
		out = append(out, k, v)
	}
	return out
}
