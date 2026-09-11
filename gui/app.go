package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pengyongshi/cutx/internal/engine"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds the Wails context and provides methods to the frontend.
type App struct {
	ctx       context.Context
	running   int32
	version   string
	reporter *wailsReporter
}

// wailsReporter implements engine.Reporter, emitting events to the frontend.
type wailsReporter struct {
	ctx       context.Context
	total     int64
	done      int64
	label     string
	lastEmit  time.Time
	mu        sync.Mutex
}

func (r *wailsReporter) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	level := "info"
	switch {
	case strings.HasPrefix(msg, "✓"):
		level = "success"
	case strings.HasPrefix(msg, "✗"):
		level = "error"
	case strings.HasPrefix(msg, "⚠"):
		level = "warning"
	}
	wailsruntime.EventsEmit(r.ctx, "log", map[string]interface{}{
		"time":    time.Now().Format("15:04:05"),
		"message": msg,
		"level":   level,
	})
}

func (r *wailsReporter) ProgressStart(total int64, label string) {
	r.mu.Lock()
	r.total = total
	r.done = 0
	r.label = label
	r.lastEmit = time.Now()
	status := r.getStatus()
	r.mu.Unlock()
	wailsruntime.EventsEmit(r.ctx, "progress", status)
}

func (r *wailsReporter) ProgressAdd(n int64) {
	r.mu.Lock()
	r.done += n
	now := time.Now()
	if now.Sub(r.lastEmit) < 200*time.Millisecond && r.done < r.total {
		r.mu.Unlock()
		return
	}
	r.lastEmit = now
	status := r.getStatus()
	r.mu.Unlock()
	wailsruntime.EventsEmit(r.ctx, "progress", status)
}

func (r *wailsReporter) ProgressFinish() {
	r.mu.Lock()
	r.done = r.total
	status := r.getStatus()
	r.mu.Unlock()
	wailsruntime.EventsEmit(r.ctx, "progress", status)
}

func (r *wailsReporter) ProgressStop() {}

func (r *wailsReporter) getStatus() map[string]interface{} {
	pct := 0.0
	if r.total > 0 {
		pct = float64(r.done) / float64(r.total) * 100
		if pct > 100 {
			pct = 100
		}
	}
	return map[string]interface{}{
		"percent":     pct,
		"label":       r.label,
		"done_bytes":  r.done,
		"total_bytes": r.total,
	}
}

var _ engine.Reporter = (*wailsReporter)(nil)

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{
		version: "v1.0.0",
	}
}

// OnStartup is called when the app starts. It saves the context and sets up
// the drag-and-drop handler.
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.reporter = &wailsReporter{ctx: ctx}

	// Handle file drops
	wailsruntime.OnFileDrop(ctx, func(x, y int, paths []string) {
		if len(paths) > 0 {
			wailsruntime.EventsEmit(ctx, "file-drop", paths[0])
		}
	})
}

// OnShutdown is called when the app is closing.
func (a *App) OnShutdown(ctx context.Context) {
	// Nothing to clean up
}

// --- Exposed methods (called from frontend) ---

// SplitRequest holds parameters for a split operation.
type SplitRequest struct {
	Source string `json:"source"`
	Size   string `json:"size"`
	Hash   string `json:"hash"`
	Output string `json:"output"`
}

// MergeRequest holds parameters for a merge operation.
type MergeRequest struct {
	Manifest string `json:"manifest"`
	Mode     string `json:"mode"`
	Output   string `json:"output"`
	Force    bool   `json:"force"`
}

// VerifyRequest holds parameters for a verify operation.
type VerifyRequest struct {
	Manifest string `json:"manifest"`
}

// GetVersion returns the application version.
func (a *App) GetVersion() string {
	return a.version
}

// Split starts a file split operation asynchronously.
func (a *App) Split(req SplitRequest) error {
	if !atomic.CompareAndSwapInt32(&a.running, 0, 1) {
		return fmt.Errorf("another operation is running")
	}
	go func() {
		defer atomic.StoreInt32(&a.running, 0)
		err := engine.Split(engine.SplitOptions{
			SourcePath: req.Source,
			ChunkSize:  req.Size,
			OutputDir:  req.Output,
			HashAlgo:   req.Hash,
			ToolVer:    a.version,
		}, a.reporter)
		if err != nil {
			a.reporter.Log("✗ Error: %v", err)
		}
		wailsruntime.EventsEmit(a.ctx, "operation-complete", map[string]interface{}{
			"success": err == nil,
		})
	}()
	return nil
}

// Merge starts a file merge operation asynchronously.
func (a *App) Merge(req MergeRequest) error {
	if !atomic.CompareAndSwapInt32(&a.running, 0, 1) {
		return fmt.Errorf("another operation is running")
	}
	go func() {
		defer atomic.StoreInt32(&a.running, 0)
		err := engine.Merge(engine.MergeOptions{
			ManifestPath: req.Manifest,
			Mode:         req.Mode,
			OutputDir:    req.Output,
			Force:        req.Force,
		}, a.reporter)
		if err != nil {
			a.reporter.Log("✗ Error: %v", err)
		}
		wailsruntime.EventsEmit(a.ctx, "operation-complete", map[string]interface{}{
			"success": err == nil,
		})
	}()
	return nil
}

// Verify starts a file verify operation asynchronously.
func (a *App) Verify(req VerifyRequest) error {
	if !atomic.CompareAndSwapInt32(&a.running, 0, 1) {
		return fmt.Errorf("another operation is running")
	}
	go func() {
		defer atomic.StoreInt32(&a.running, 0)
		err := engine.Verify(engine.VerifyOptions{
			ManifestPath: req.Manifest,
		}, a.reporter)
		if err != nil {
			a.reporter.Log("✗ Error: %v", err)
		}
		wailsruntime.EventsEmit(a.ctx, "operation-complete", map[string]interface{}{
			"success": err == nil,
		})
	}()
	return nil
}

// WindowMinimise minimizes the window (called from custom title bar).
func (a *App) WindowMinimise() {
	wailsruntime.WindowMinimise(a.ctx)
}

// WindowClose closes the window (called from custom title bar).
func (a *App) WindowClose() {
	wailsruntime.Quit(a.ctx)
}
