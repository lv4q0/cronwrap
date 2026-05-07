package metrics

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// Reporter writes a human-readable metrics summary to an io.Writer.
type Reporter struct {
	w io.Writer
}

// NewReporter creates a Reporter that writes to w.
func NewReporter(w io.Writer) *Reporter {
	return &Reporter{w: w}
}

// Print formats and writes the summary snapshot to the reporter's writer.
func (r *Reporter) Print(s *Summary) error {
	snap := s.Snapshot()

	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "METRIC\tVALUE")
	fmt.Fprintln(tw, "------\t-----")
	fmt.Fprintf(tw, "Total Runs\t%d\n", snap.TotalRuns)
	fmt.Fprintf(tw, "Success Runs\t%d\n", snap.SuccessRuns)
	fmt.Fprintf(tw, "Failure Runs\t%d\n", snap.FailureRuns)
	fmt.Fprintf(tw, "Total Retries\t%d\n", snap.TotalRetries)
	fmt.Fprintf(tw, "Success Rate\t%.1f%%\n", snap.SuccessRate()*100)
	fmt.Fprintf(tw, "Last Run At\t%s\n", formatTime(snap.LastRunAt))
	fmt.Fprintf(tw, "Last Duration\t%s\n", snap.LastDuration.Round(time.Millisecond))
	fmt.Fprintf(tw, "Last Status\t%s\n", statusLabel(snap.LastSuccess))

	return tw.Flush()
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format(time.RFC3339)
}

func statusLabel(ok bool) string {
	if ok {
		return "success"
	}
	return "failure"
}
