// Package backoff provides retry delay strategies for use with the runner.
package backoff

import (
	"math"
	"time"
)

// Strategy defines how long to wait before the next retry attempt.
type Strategy interface {
	// Next returns the delay before attempt number n (0-indexed).
	Next(attempt int) time.Duration
}

// ConstantStrategy waits the same duration between every retry.
type ConstantStrategy struct {
	Delay time.Duration
}

func (c ConstantStrategy) Next(_ int) time.Duration {
	return c.Delay
}

// ExponentialStrategy implements truncated binary exponential back-off.
//
//	delay = Base * 2^attempt   (capped at Max)
type ExponentialStrategy struct {
	Base time.Duration
	Max  time.Duration
}

func (e ExponentialStrategy) Next(attempt int) time.Duration {
	if e.Base <= 0 {
		e.Base = time.Second
	}
	multiplier := math.Pow(2, float64(attempt))
	d := time.Duration(float64(e.Base) * multiplier)
	if e.Max > 0 && d > e.Max {
		return e.Max
	}
	return d
}

// LinearStrategy increases the delay linearly with each attempt.
//
//	delay = Base + Step*attempt
type LinearStrategy struct {
	Base time.Duration
	Step time.Duration
}

func (l LinearStrategy) Next(attempt int) time.Duration {
	return l.Base + l.Step*time.Duration(attempt)
}

// Default returns the default exponential strategy used by the runner
// when no explicit strategy is provided.
func Default() Strategy {
	return ExponentialStrategy{
		Base: time.Second,
		Max:  30 * time.Second,
	}
}
