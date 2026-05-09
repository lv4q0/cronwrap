package middleware

import (
	"context"
	"fmt"
)

// RecoverMiddleware catches any panic produced by the wrapped RunFunc and
// converts it into a non-nil error so the runner can handle it gracefully.
func RecoverMiddleware() Middleware {
	return func(next RunFunc) RunFunc {
		return func(ctx context.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("job panicked: %v", r)
				}
			}()
			return next(ctx)
		}
	}
}
