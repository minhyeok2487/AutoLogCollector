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
	Servers        []Server
	Commands       []string
	Credentials    *Credentials
	LogDir         string
	MaxConcurrent  int  // Maximum number of concurrent executions (always 1)
	ChunkTimeout   int  // Seconds to wait for data chunks
	EnableMode     bool // Whether to enter enable mode
	DisablePaging  bool // Whether to disable paging (terminal length 0)
	OnProgress     ProgressCallback
	OnResult       ResultCallback
	OnLog          LogCallback // Real-time log callback
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.Mutex
	isRunning      bool
	results        []ExecutionResult
	successCount   int
	failCount      int
	completedCount int
}

// NewRunner creates a new Runner instance
func NewRunner(servers []Server, commands []string, creds *Credentials, chunkTimeout int, enableMode, disablePaging bool, scheduleName string) *Runner {
	date := time.Now().Format("2006-01-02")
	var logDir string
	if scheduleName != "" {
		logDir = filepath.Join("logs", scheduleName, date)
	} else {
		logDir = filepath.Join("logs", date)
	}
	if chunkTimeout <= 0 {
		chunkTimeout = 1
	}
	return &Runner{
		Servers:       servers,
		Commands:      commands,
		Credentials:   creds,
		LogDir:        logDir,
		MaxConcurrent: 1, // Force sequential execution to avoid output truncation issues
		ChunkTimeout:  chunkTimeout,
		EnableMode:    enableMode,
		DisablePaging: disablePaging,
		results:       make([]ExecutionResult, 0, len(servers)),
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
	r.completedCount = 0
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

// serverJob represents a job to process a server
type serverJob struct {
	index  int
	server Server
}

func (r *Runner) run() {
	defer func() {
		r.mu.Lock()
		r.isRunning = false
		r.mu.Unlock()
	}()

	// Create job channel
	jobs := make(chan serverJob, len(r.Servers))

	// Create wait group for workers
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < r.MaxConcurrent; w++ {
		wg.Add(1)
		go r.worker(&wg, jobs)
	}

	// Send jobs
	for i, server := range r.Servers {
		select {
		case <-r.ctx.Done():
			close(jobs)
			wg.Wait()
			return
		case jobs <- serverJob{index: i, server: server}:
		}
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()
}

func (r *Runner) worker(wg *sync.WaitGroup, jobs <-chan serverJob) {
	defer wg.Done()

	for job := range jobs {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		server := job.server

		if r.OnProgress != nil {
			r.mu.Lock()
			completed := r.completedCount
			r.mu.Unlock()
			r.OnProgress(completed+1, len(r.Servers), server, "connecting")
		}

		startTime := time.Now()
		result := ExecutionResult{
			Server: server,
		}

		// Create log callback for this server
		var logCallback func(line string)
		if r.OnLog != nil {
			logCallback = func(line string) {
				r.OnLog(server.IP, server.Hostname, line)
			}
		}

		// Use per-server credentials if set, otherwise use global credentials
		creds := r.Credentials
		if server.Username != "" && server.Password != "" {
			creds = &Credentials{
				User:           server.Username,
				Password:       server.Password,
				EnablePassword: server.EnablePassword,
			}
		}

		output, err := ExecuteCommands(server, creds, r.Commands, r.ChunkTimeout, r.EnableMode, r.DisablePaging, logCallback)
		result.Duration = time.Since(startTime).Milliseconds()

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			r.mu.Lock()
			r.failCount++
			r.completedCount++
			r.mu.Unlock()
		} else {
			// Save log
			logPath := filepath.Join(r.LogDir, server.Hostname+".log")
			if saveErr := SaveLog(logPath, output); saveErr != nil {
				result.Success = false
				result.Error = "Failed to save log: " + saveErr.Error()
				r.mu.Lock()
				r.failCount++
				r.completedCount++
				r.mu.Unlock()
			} else {
				result.Success = true
				result.Output = output
				result.LogPath = logPath
				r.mu.Lock()
				r.successCount++
				r.completedCount++
				r.mu.Unlock()
			}
		}

		r.mu.Lock()
		r.results = append(r.results, result)
		completed := r.completedCount
		r.mu.Unlock()

		if r.OnResult != nil {
			r.OnResult(result)
		}

		if r.OnProgress != nil {
			status := "success"
			if !result.Success {
				status = "failed"
			}
			r.OnProgress(completed, len(r.Servers), server, status)
		}
	}
}
