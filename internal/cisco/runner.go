package cisco

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Runner orchestrates command execution across multiple servers
type Runner struct {
	Servers      []Server
	Commands     []string
	Credentials  *Credentials
	LogDir       string
	OnProgress   ProgressCallback
	OnResult     ResultCallback
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	isRunning    bool
	results      []ExecutionResult
	successCount int
	failCount    int
}

// NewRunner creates a new Runner instance
func NewRunner(servers []Server, commands []string, creds *Credentials) *Runner {
	logDir := filepath.Join("logs", time.Now().Format("2006-01-02"))
	return &Runner{
		Servers:     servers,
		Commands:    commands,
		Credentials: creds,
		LogDir:      logDir,
		results:     make([]ExecutionResult, 0, len(servers)),
	}
}

// Start begins the execution process
func (r *Runner) Start() error {
	r.mu.Lock()
	if r.isRunning {
		r.mu.Unlock()
		return nil
	}
	r.isRunning = true
	r.results = make([]ExecutionResult, 0, len(r.Servers))
	r.successCount = 0
	r.failCount = 0
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.mu.Unlock()

	// Create log directory
	if err := os.MkdirAll(r.LogDir, 0755); err != nil {
		r.mu.Lock()
		r.isRunning = false
		r.mu.Unlock()
		return err
	}

	go r.run()
	return nil
}

// Stop cancels the running execution
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
}

// IsRunning returns whether the runner is currently executing
func (r *Runner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isRunning
}

// GetResults returns the current results
func (r *Runner) GetResults() []ExecutionResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]ExecutionResult{}, r.results...)
}

// GetSummary returns success and fail counts
func (r *Runner) GetSummary() (success, fail, total int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.successCount, r.failCount, len(r.Servers)
}

func (r *Runner) run() {
	defer func() {
		r.mu.Lock()
		r.isRunning = false
		r.mu.Unlock()
	}()

	for i, server := range r.Servers {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		if r.OnProgress != nil {
			r.OnProgress(i+1, len(r.Servers), server, "connecting")
		}

		startTime := time.Now()
		result := ExecutionResult{
			Server: server,
		}

		output, err := ExecuteCommands(server, r.Credentials, r.Commands)
		result.Duration = time.Since(startTime).Milliseconds()

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			r.mu.Lock()
			r.failCount++
			r.mu.Unlock()
		} else {
			// Save log
			logPath := filepath.Join(r.LogDir, server.Hostname+".log")
			if saveErr := SaveLog(logPath, output); saveErr != nil {
				result.Success = false
				result.Error = "Failed to save log: " + saveErr.Error()
				r.mu.Lock()
				r.failCount++
				r.mu.Unlock()
			} else {
				result.Success = true
				result.Output = output
				result.LogPath = logPath
				r.mu.Lock()
				r.successCount++
				r.mu.Unlock()
			}
		}

		r.mu.Lock()
		r.results = append(r.results, result)
		r.mu.Unlock()

		if r.OnResult != nil {
			r.OnResult(result)
		}

		if r.OnProgress != nil {
			status := "success"
			if !result.Success {
				status = "failed"
			}
			r.OnProgress(i+1, len(r.Servers), server, status)
		}
	}
}
