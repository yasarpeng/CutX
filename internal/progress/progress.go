package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Bar is a streaming progress bar with ETA calculation.
type Bar struct {
	total     int64
	done      int64
	startTime time.Time
	out      io.Writer
	quiet     bool
	label     string
	lastUpd   time.Time
	mu        sync.Mutex
	stop      atomic.Bool
}

// New creates a progress bar.
func New(total int64, label string, quiet bool) *Bar {
	return &Bar{
		total:     total,
		startTime: time.Now(),
		out:        os.Stderr,
		quiet:      quiet,
		label:      label,
	}
}

// Add increments the progress by n bytes.
func (b *Bar) Add(n int64) {
	cur := atomic.AddInt64(&b.done, n)
	if b.quiet {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	if now.Sub(b.lastUpd) < 500*time.Millisecond && cur < b.total {
		return
	}
	b.lastUpd = now
	b.render(cur)
}

// Finish prints the final state.
func (b *Bar) Finish() {
	if b.quiet {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	cur := atomic.LoadInt64(&b.done)
	b.render(cur)
	fmt.Fprintf(b.out, "\n")
}

// Elapsed returns the elapsed time since start.
func (b *Bar) Elapsed() time.Duration {
	return time.Since(b.startTime)
}

// Stop signals the bar to stop rendering (e.g. on interrupt).
func (b *Bar) Stop() {
	b.stop.Store(true)
}

func (b *Bar) render(done int64) {
	if b.total <= 0 {
		return
	}
	pct := float64(done) / float64(b.total) * 100
	if pct > 100 {
		pct = 100
	}
	barLen := 20
	filled := int(pct / 100 * float64(barLen))
	if filled > barLen {
		filled = barLen
	}
	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", barLen-filled)

	elapsed := time.Since(b.startTime).Seconds()
	var eta string
	if done > 0 && elapsed > 5 {
		speed := float64(done) / elapsed
		remaining := float64(b.total-done) / speed
		eta = formatDuration(remaining)
	} else if done > 0 {
		eta = "计算中..."
	} else {
		eta = "--"
	}

	fmt.Fprintf(b.out, "\r[ cutx ] %s %s %.1f%% | %s / %s | ETA: %s   ",
		b.label, bar, pct, formatBytes(done), formatBytes(b.total), eta)
}

func formatBytes(b int64) string {
	const (
		GB = 1024 * 1024 * 1024
		MB = 1024 * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1fGB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1fMB", float64(b)/float64(MB))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

func formatDuration(seconds float64) string {
	if seconds < 0 {
		return "--"
	}
	m := int(seconds) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%dm%02ds", m, s)
}
