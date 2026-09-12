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

type App struct {
	ctx       context.Context
	running   int32
	version   string
	reporter *wailsReporter
}

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

func NewApp() *App {
	return &App{version: "v1.0.0"}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.reporter = &wailsReporter{ctx: ctx}
	wailsruntime.OnFileDrop(ctx, func(x, y int, paths []string) {
		if len(paths) > 0 {
			wailsruntime.EventsEmit(ctx, "file-drop", paths[0])
		}
	})
}

func (a *App) OnShutdown(ctx context.Context) {}

// --- Exposed methods ---

type SplitRequest struct {
	Source string `json:"source"`
	Size   string `json:"size"`
	Hash   string `json:"hash"`
	Output string `json:"output"`
}

type MergeRequest struct {
	Manifest string `json:"manifest"`
	Mode     string `json:"mode"`
	Output   string `json:"output"`
	Force    bool   `json:"force"`
}

type VerifyRequest struct {
	Manifest string `json:"manifest"`
}

func (a *App) GetVersion() string {
	return a.version
}

func (a *App) OpenFileDialog(title, filter string) (string, error) {
	filters := []wailsruntime.FileFilter{
		{DisplayName: "All files", Pattern: "*.*"},
	}
	if strings.Contains(filter, "manifest") {
		filters = []wailsruntime.FileFilter{
			{DisplayName: "Manifest files", Pattern: "*.manifest.json"},
			{DisplayName: "JSON files", Pattern: "*.json"},
			{DisplayName: "All files", Pattern: "*.*"},
		}
	}
	return wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:   title,
		Filters: filters,
	})
}

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

func (a *App) WindowMinimise() {
	wailsruntime.WindowMinimise(a.ctx)
}

func (a *App) WindowClose() {
	wailsruntime.Quit(a.ctx)
}
