// Package circuit provides a circuit breaker for cron job execution.
//
// A Breaker tracks consecutive failures for a job. Once the failure count
// reaches a configurable threshold the circuit "opens" and all subsequent
// execution attempts are blocked until a cooldown period has elapsed.
//
// After the cooldown the circuit enters a "half-open" state: a single probe
// run is permitted. A successful probe closes the circuit and resets the
// failure counter; another failure reopens it and restarts the cooldown.
//
// Usage via Guard (recommended):
//
//	g := circuit.NewGuard("backup", 3, 5*time.Minute, log)
//	err := g.Execute(ctx, func(ctx context.Context) error {
//	    return runBackup(ctx)
//	})
//
// Usage via Breaker directly:
//
//	b := circuit.New(3, 5*time.Minute)
//	if ok, _ := b.Allow(); ok {
//	    if err := run(); err != nil {
//	        b.RecordFailure()
//	    } else {
//	        b.RecordSuccess()
//	    }
//	}
package circuit
