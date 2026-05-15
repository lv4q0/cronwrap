package middleware

import (
	"context"
	"os"
	"strings"
)

type envKey struct{}

// EnvMiddleware injects a set of environment variable key=value pairs into the
// job's context and optionally sets them in the process environment for the
// duration of the run.
func EnvMiddleware(vars map[string]string, setProcess bool) func(JobFunc) JobFunc {
	return func(next JobFunc) JobFunc {
		return func(ctx context.Context) error {
			ctx = context.WithValue(ctx, envKey{}, vars)

			if setProcess {
				originals := make(map[string]string, len(vars))
				for k, v := range vars {
					originals[k] = os.Getenv(k)
					os.Setenv(k, v)
				}
				defer func() {
					for k, orig := range originals {
						if orig == "" {
							os.Unsetenv(k)
						} else {
							os.Setenv(k, orig)
						}
					}
				}()
			}

			return next(ctx)
		}
	}
}

// EnvFromContext returns the env vars injected by EnvMiddleware, or nil.
func EnvFromContext(ctx context.Context) map[string]string {
	v, _ := ctx.Value(envKey{}).(map[string]string)
	return v
}

// FormatEnv returns a slice of "KEY=VALUE" strings suitable for exec.Cmd.Env.
func FormatEnv(vars map[string]string) []string {
	out := make([]string, 0, len(vars))
	for k, v := range vars {
		out = append(out, strings.ToUpper(k)+"="+v)
	}
	return out
}
