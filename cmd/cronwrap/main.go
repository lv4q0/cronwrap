// Package main is the entry point for the cronwrap CLI.
// It loads a configuration file, wires together all middleware layers,
// and either runs the job once or schedules it via cron depending on
// the flags provided.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/cronwrap/internal/alert"
	"github.com/yourorg/cronwrap/internal/audit"
	"github.com/yourorg/cronwrap/internal/backoff"
	"github.com/yourorg/cronwrap/internal/circuit"
	"github.com/yourorg/cronwrap/internal/config"
	"github.com/yourorg/cronwrap/internal/history"
	"github.com/yourorg/cronwrap/internal/logger"
	"github.com/yourorg/cronwrap/internal/metrics"
	"github.com/yourorg/cronwrap/internal/middleware"
	"github.com/yourorg/cronwrap/internal/notify"
	"github.com/yourorg/cronwrap/internal/ratelimit"
	"github.com/yourorg/cronwrap/internal/runner"
	"github.com/yourorg/cronwrap/internal/schedule"
	"github.com/yourorg/cronwrap/internal/webhook"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "cronwrap: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		cfgPath  = flag.String("config", "cronwrap.yaml", "path to configuration file")
		runOnce  = flag.Bool("once", false, "run the job once and exit instead of scheduling")
		verbose  = flag.Bool("verbose", false, "enable debug-level logging")
		webhookAddr = flag.String("webhook-addr", "", "address for the status/metrics HTTP server (e.g. :8080)")
	)
	flag.Parse()

	cfg, err := config.LoadFile(*cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log := logger.NewDefault()
	if *verbose {
		log = logger.New(logger.DebugLevel, os.Stdout)
	}

	// --- metrics & history -------------------------------------------------
	store := metrics.NewStore()
	hisStore, err := history.NewStore(cfg.HistoryPath)
	if err != nil {
		return fmt.Errorf("opening history store: %w", err)
	}

	// --- audit log ---------------------------------------------------------
	auditLog, err := audit.NewLog(cfg.AuditPath)
	if err != nil {
		return fmt.Errorf("opening audit log: %w", err)
	}

	// --- alerting / notifications ------------------------------------------
	var senders []notify.Sender
	if cfg.WebhookURL != "" {
		senders = append(senders, alert.NewWebhookNotifier(cfg.WebhookURL))
	}
	dispatcher := notify.NewDispatcher(senders, notify.Info)

	// --- circuit breaker ---------------------------------------------------
	breaker := circuit.New(circuit.Options{
		FailureThreshold: cfg.CircuitBreaker.FailureThreshold,
		Cooldown:         time.Duration(cfg.CircuitBreaker.CooldownSeconds) * time.Second,
	})

	// --- rate-limit store --------------------------------------------------
	rlStore := ratelimit.NewStore()

	// --- middleware chain --------------------------------------------------
	chain := middleware.Chain(
		middleware.RecoverMiddleware,
		middleware.LogMiddleware(log),
		middleware.NewMetricsMiddleware(store),
		audit.NewMiddleware(auditLog, cfg.Name),
		middleware.NewNotifyMiddleware(dispatcher, cfg.Name),
		middleware.CircuitMiddleware(breaker, log),
		middleware.ThrottleMiddleware(rlStore, cfg.Name,
			time.Duration(cfg.RateLimit.IntervalSeconds)*time.Second, nil),
		middleware.RetryMiddleware(cfg.Retry.MaxAttempts,
			backoff.Default()),
		middleware.TimeoutMiddleware(time.Duration(cfg.TimeoutSeconds)*time.Second),
	)

	// --- optional HTTP status server ---------------------------------------
	if *webhookAddr != "" {
		srv := webhook.NewServer(*webhookAddr,
			webhook.New(store, hisStore))
		go func() {
			if srvErr := srv.Start(); srvErr != nil {
				log.Error("webhook server error", map[string]any{"error": srvErr})
			}
		}()
		defer srv.Shutdown(context.Background()) //nolint:errcheck
	}

	// --- run once or on schedule ------------------------------------------
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	opts := runner.DefaultOptions()
	opts.Command = cfg.Command
	opts.Args = cfg.Args
	opts.Middleware = chain

	if *runOnce {
		return runner.Run(ctx, opts)
	}

	sched, err := schedule.Parse(cfg.Schedule)
	if err != nil {
		return fmt.Errorf("parsing schedule %q: %w", cfg.Schedule, err)
	}

	log.Info("cronwrap started", map[string]any{
		"job":      cfg.Name,
		"schedule": cfg.Schedule,
	})

	for {
		next := sched.Next(time.Now())
		select {
		case <-ctx.Done():
			log.Info("cronwrap shutting down", nil)
			return nil
		case <-time.After(time.Until(next)):
			if runErr := runner.Run(ctx, opts); runErr != nil {
				log.Error("job failed", map[string]any{"error": runErr})
			}
		}
	}
}
