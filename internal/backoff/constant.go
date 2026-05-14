package backoff

import "time"

// constantStrategy returns the same delay for every attempt.
type constantStrategy struct {
	delay time.Duration
}

// Constant returns a Strategy that always waits the given duration.
// Pass 0 to skip any delay (useful in tests).
func Constant(d time.Duration) Strategy {
	return &constantStrategy{delay: d}
}

func (c *constantStrategy) Delay(_ int) time.Duration {
	return c.delay
}
