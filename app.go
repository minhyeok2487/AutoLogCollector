package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"cisco-plink/internal/cisco"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx           context.Context
	runner        *cisco.Runner
	mu            sync.Mutex
	serversFile   string
	commandsFile  string
	servers       []cisco.Server
	commands      []string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectServersFile opens a file dialog to select servers CSV file
func (a *App) SelectServersFile() string {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Servers CSV File",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil || file == "" {
		return ""
	}
	a.serversFile = file
	return file
}

// SelectCommandsFile opens a file dialog to select commands file
func (a *App) SelectCommandsFile() string {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Commands File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil || file == "" {
		return ""
	}
	a.commandsFile = file
	return file
}

// PreviewServers loads and returns the list of servers from the selected file
func (a *App) PreviewServers() []cisco.Server {
	if a.serversFile == "" {
		return nil
	}
	servers, err := cisco.LoadServers(a.serversFile)
	if err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to load servers: "+err.Error())
		return nil
	}
	a.servers = servers
	return servers
}

// PreviewCommands loads and returns the commands from the selected file
func (a *App) PreviewCommands() []string {
	if a.commandsFile == "" {
		return nil
	}
	commands, err := cisco.LoadCommands(a.commandsFile)
	if err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to load commands: "+err.Error())
		return nil
	}
	a.commands = commands
	return commands
}

// StartExecution begins the command execution
func (a *App) StartExecution(username, password string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.runner != nil && a.runner.IsRunning() {
		return false
	}

	if len(a.servers) == 0 {
		runtime.EventsEmit(a.ctx, "error", "No servers loaded")
		return false
	}

	if len(a.commands) == 0 {
		runtime.EventsEmit(a.ctx, "error", "No commands loaded")
		return false
	}

	if username == "" || password == "" {
		runtime.EventsEmit(a.ctx, "error", "Username and password are required")
		return false
	}

	creds := &cisco.Credentials{
		User:     username,
		Password: password,
	}

	a.runner = cisco.NewRunner(a.servers, a.commands, creds)

	// Set progress callback
	a.runner.OnProgress = func(current, total int, server cisco.Server, status string) {
		runtime.EventsEmit(a.ctx, "progress", map[string]interface{}{
			"current":  current,
			"total":    total,
			"hostname": server.Hostname,
			"ip":       server.IP,
			"status":   status,
		})
	}

	// Set result callback
	a.runner.OnResult = func(result cisco.ExecutionResult) {
		// Convert backslashes to forward slashes for JavaScript compatibility
		logPath := strings.ReplaceAll(result.LogPath, "\\", "/")
		runtime.EventsEmit(a.ctx, "result", map[string]interface{}{
			"hostname": result.Server.Hostname,
			"ip":       result.Server.IP,
			"success":  result.Success,
			"error":    result.Error,
			"logPath":  logPath,
			"duration": result.Duration,
		})

		// Check if all done
		success, fail, total := a.runner.GetSummary()
		if success+fail == total {
			runtime.EventsEmit(a.ctx, "completed", map[string]interface{}{
				"success": success,
				"fail":    fail,
				"total":   total,
				"logDir":  a.runner.LogDir,
			})
		}
	}

	if err := a.runner.Start(); err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to start: "+err.Error())
		return false
	}

	return true
}

// StopExecution stops the running execution
func (a *App) StopExecution() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.runner != nil {
		a.runner.Stop()
	}
}

// IsRunning returns whether execution is in progress
func (a *App) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.runner == nil {
		return false
	}
	return a.runner.IsRunning()
}

// OpenLogsFolder opens the logs folder in file explorer
func (a *App) OpenLogsFolder() {
	logsDir := "logs"
	if a.runner != nil && a.runner.LogDir != "" {
		logsDir = a.runner.LogDir
	}

	// Create logs directory if it doesn't exist
	os.MkdirAll(logsDir, 0755)

	// Get absolute path
	absPath, err := filepath.Abs(logsDir)
	if err != nil {
		absPath = logsDir
	}

	// Open folder using explorer.exe on Windows
	exec.Command("explorer.exe", absPath).Start()
}

// GetLogFiles returns list of log files in the logs directory
func (a *App) GetLogFiles() []map[string]string {
	logsDir := "logs"
	var files []map[string]string

	// Walk through logs directory
	filepath.Walk(logsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".log" {
			// Convert backslashes to forward slashes for JavaScript compatibility
			normalizedPath := strings.ReplaceAll(path, "\\", "/")
			files = append(files, map[string]string{
				"name":    info.Name(),
				"path":    normalizedPath,
				"date":    filepath.Base(filepath.Dir(path)),
				"modTime": info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
		return nil
	})

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i]["modTime"] > files[j]["modTime"]
	})

	return files
}

// ReadLogFile reads and returns the content of a log file
func (a *App) ReadLogFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return "Error reading file: " + err.Error()
	}
	return string(content)
}

// GetCurrentLogDir returns the current log directory
func (a *App) GetCurrentLogDir() string {
	if a.runner != nil {
		return a.runner.LogDir
	}
	return ""
}
