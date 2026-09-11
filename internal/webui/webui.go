package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/pengyongshi/cutx/internal/engine"
)

//go:embed static/index.html
var staticFiles embed.FS

type webReporter struct {
	mu      sync.Mutex
	logs    []logEntry
	current string
	total   int64
	done    int64
	running bool
}

type logEntry struct {
	Time   string `json:"time"`
	Msg    string `json:"msg"`
	Level  string `json:"level"`
}

func (r *webReporter) Log(format string, args ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, logEntry{
		Time:  time.Now().Format("15:04:05"),
		Msg:   fmt.Sprintf(format, args...),
		Level: "info",
	})
}

func (r *webReporter) ProgressStart(total int64, label string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.total = total
	r.done = 0
	r.current = label
}

func (r *webReporter) ProgressAdd(n int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.done += n
}

func (r *webReporter) ProgressFinish() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.done = r.total
}

func (r *webReporter) ProgressStop() {}

var _ engine.Reporter = (*webReporter)(nil)

type server struct {
	rep     *webReporter
	version string
	commit  string
}

// Run starts the web UI server and opens the browser.
// This is called when cutx is run with no arguments.
func Run(ver, com string) {
	rep := &webReporter{}
	srv := &server{rep: rep, version: ver, commit: com}

	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()

	// Serve the embedded HTML
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, _ := staticFiles.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// API endpoints
	mux.HandleFunc("/api/version", srv.handleVersion)
	mux.HandleFunc("/api/split", srv.handleSplit)
	mux.HandleFunc("/api/merge", srv.handleMerge)
	mux.HandleFunc("/api/verify", srv.handleVerify)
	mux.HandleFunc("/api/status", srv.handleStatus)

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Printf("CutX GUI starting at %s\n", url)

	// Open browser after a short delay
	go func() {
		time.Sleep(300 * time.Millisecond)
		openBrowser(url)
	}()

	// Start server (blocks)
	if err := http.Serve(listener, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func (s *server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"version": s.version})
}

func (s *server) handleSplit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		Source string `json:"source"`
		Size   string `json:"size"`
		Hash   string `json:"hash"`
		Output string `json:"output"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	s.rep.running = true
	go func() {
		err := engine.Split(engine.SplitOptions{
			SourcePath: req.Source,
			ChunkSize:  req.Size,
			OutputDir:  req.Output,
			HashAlgo:  req.Hash,
			ToolVer:   s.version + " (" + s.commit + ")",
		}, s.rep)
		if err != nil {
			s.rep.Log("✗ Error: %v", err)
		}
		s.rep.running = false
	}()

	writeJSON(w, map[string]string{"status": "started"})
}

func (s *server) handleMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		Manifest string `json:"manifest"`
		Mode     string `json:"mode"`
		Output   string `json:"output"`
		Force    bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	s.rep.running = true
	go func() {
		err := engine.Merge(engine.MergeOptions{
			ManifestPath: req.Manifest,
			Mode:         req.Mode,
			OutputDir:    req.Output,
			Force:        req.Force,
		}, s.rep)
		if err != nil {
			s.rep.Log("✗ Error: %v", err)
		}
		s.rep.running = false
	}()

	writeJSON(w, map[string]string{"status": "started"})
}

func (s *server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		Manifest string `json:"manifest"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	s.rep.running = true
	go func() {
		err := engine.Verify(engine.VerifyOptions{
			ManifestPath: req.Manifest,
		}, s.rep)
		if err != nil {
			s.rep.Log("✗ Error: %v", err)
		}
		s.rep.running = false
	}()

	writeJSON(w, map[string]string{"status": "started"})
}

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.rep.mu.Lock()
	defer s.rep.mu.Unlock()

	pct := 0.0
	if s.rep.total > 0 {
		pct = float64(s.rep.done) / float64(s.rep.total) * 100
		if pct > 100 {
			pct = 100
		}
	}

	resp := map[string]interface{}{
		"progress":   pct,
		"status":     s.rep.current,
		"running":    s.rep.running,
		"logs":       s.rep.logs,
	}
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", url)
		fmt.Printf("Please open %s manually\n", url)
	}
}
