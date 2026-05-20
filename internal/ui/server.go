// File server.go hosts the interactive local cleanup UI.
package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"digital-exhaust-cleaner/internal/analyzer"
	"digital-exhaust-cleaner/internal/cleanup"
)

const defaultServerTimeout = 15 * time.Second

// ServerConfig defines the local interactive UI server.
type ServerConfig struct {
	Addr          string
	Root          string
	QuarantineDir string
	Result        analyzer.Result
	ScanFunc      func(ctx context.Context, path string) (analyzer.Result, error)
}

// Serve starts a loopback-only interactive cleanup server.
func Serve(ctx context.Context, cfg ServerConfig) error {
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:8787"
	}
	if err := ensureLoopback(cfg.Addr); err != nil {
		return err
	}

	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return fmt.Errorf("resolve server root: %w", err)
	}

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: defaultServerTimeout,
	}
	manager := cleanup.NewManager(cfg.QuarantineDir)

	// Serve compiled frontend assets (CSS + JS) embedded in the binary.
	staticContent, err := fs.Sub(StaticFS, "static")
	if err != nil {
		return fmt.Errorf("mount static assets: %w", err)
	}
	staticHandler := http.FileServer(http.FS(staticContent))
	mux.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Instruct browsers to cache compiled assets aggressively.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		staticHandler.ServeHTTP(w, r)
	})))

	var stateMutex sync.Mutex
	currentResult := cfg.Result
	currentRoot := root

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		stateMutex.Lock()
		defer stateMutex.Unlock()

		targetPath := r.URL.Query().Get("path")
		if targetPath != "" && targetPath != currentRoot && cfg.ScanFunc != nil {
			absPath, err := filepath.Abs(targetPath)
			if err == nil {
				res, err := cfg.ScanFunc(r.Context(), absPath)
				if err != nil {
					http.Error(w, fmt.Sprintf("scan failed: %v", err), http.StatusInternalServerError)
					return
				}
				currentResult = res
				currentRoot = absPath
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := reportTemplate.Execute(w, viewModel{Result: currentResult, Interactive: true}); err != nil {
			http.Error(w, "render report", http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/api/quarantine", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		target, err := filepath.Abs(request.Path)
		if err != nil {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		
		stateMutex.Lock()
		rootContext := currentRoot
		stateMutex.Unlock()

		if !isInside(rootContext, target) {
			http.Error(w, "path is outside the scanned root", http.StatusBadRequest)
			return
		}

		record, err := manager.Quarantine(target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(record); err != nil {
			http.Error(w, "encode response", http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		history, err := manager.History()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(history); err != nil {
			http.Error(w, "encode response", http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/api/browse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path, err := selectFolderDialog()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to open folder explorer: %v", err), http.StatusInternalServerError)
			return
		}

		if path == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		response := struct {
			Path string `json:"path"`
		}{
			Path: path,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultServerTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func ensureLoopback(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("parse server address: %w", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("interactive UI must bind to a loopback address")
	}
	return nil
}

func isInside(root string, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel))
}

func selectFolderDialog() (string, error) {
	switch runtime.GOOS {
	case "windows":
		script := `
Add-Type -AssemblyName System.Windows.Forms
$f = New-Object System.Windows.Forms.FolderBrowserDialog
$f.Description = "Select a folder to scan"
$f.ShowNewFolderButton = $true
if ($f.ShowDialog() -eq "OK") {
    Write-Output $f.SelectedPath
}
`
		tmpFile, err := os.CreateTemp("", "select_folder_*.ps1")
		if err != nil {
			return "", err
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(script); err != nil {
			tmpFile.Close()
			return "", err
		}
		tmpFile.Close()

		cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", tmpFile.Name())
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("powershell folder picker failed: %v (stderr: %q)", err, stderr.String())
		}
		return strings.TrimSpace(stdout.String()), nil

	case "darwin":
		cmd := exec.Command("osascript", "-e", `POSIX path of (choose folder with prompt "Select a folder to scan")`)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		if err := cmd.Run(); err != nil {
			// Cancelled or errored
			return "", nil
		}
		return strings.TrimSpace(stdout.String()), nil

	case "linux":
		if _, err := exec.LookPath("zenity"); err == nil {
			cmd := exec.Command("zenity", "--file-selection", "--directory", "--title=Select a folder to scan")
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			if err := cmd.Run(); err != nil {
				return "", nil
			}
			return strings.TrimSpace(stdout.String()), nil
		} else if _, err := exec.LookPath("kdialog"); err == nil {
			cmd := exec.Command("kdialog", "--getexistingdirectory", ".", "--title", "Select a folder to scan")
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			if err := cmd.Run(); err != nil {
				return "", nil
			}
			return strings.TrimSpace(stdout.String()), nil
		}
		return "", fmt.Errorf("no supported dialog tool found (install zenity or kdialog)")

	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
