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
	"cisco-plink/internal/updater"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GitHub repository settings for auto-update
const (
	GitHubOwner = "your-username"
	GitHubRepo  = "cisco-plink"
)

// App struct
type App struct {
	ctx      context.Context
	runner   *cisco.Runner
	updater  *updater.Updater
	mu       sync.Mutex
	servers  []cisco.Server
	commands []string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.updater = updater.NewUpdater(GitHubOwner, GitHubRepo)
}

// SetServers sets the server list from GUI input
func (a *App) SetServers(servers []map[string]string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.servers = make([]cisco.Server, 0, len(servers))
	for _, s := range servers {
		ip := s["ip"]
		hostname := s["hostname"]
		if ip != "" {
			if hostname == "" {
				hostname = ip
			}
			a.servers = append(a.servers, cisco.Server{
				IP:       ip,
				Hostname: hostname,
			})
		}
	}
}

// SetCommands sets the command list from GUI input
func (a *App) SetCommands(commands []string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.commands = make([]string, 0, len(commands))
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			a.commands = append(a.commands, cmd)
		}
	}
}

// ImportServersFromCSV opens file dialog and returns parsed servers
func (a *App) ImportServersFromCSV() []map[string]string {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import Servers from CSV",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil || file == "" {
		return nil
	}

	servers, err := cisco.LoadServers(file)
	if err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to load CSV: "+err.Error())
		return nil
	}

	result := make([]map[string]string, len(servers))
	for i, s := range servers {
		result[i] = map[string]string{
			"ip":       s.IP,
			"hostname": s.Hostname,
		}
	}
	return result
}

// StartExecution begins the command execution
func (a *App) StartExecution(username, password string, timeout int, enableMode, disablePaging bool) bool {
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

	if timeout <= 0 {
		timeout = 1
	}

	creds := &cisco.Credentials{
		User:     username,
		Password: password,
	}

	a.runner = cisco.NewRunner(a.servers, a.commands, creds, timeout, enableMode, disablePaging)

	a.runner.OnProgress = func(current, total int, server cisco.Server, status string) {
		runtime.EventsEmit(a.ctx, "progress", map[string]interface{}{
			"current":  current,
			"total":    total,
			"hostname": server.Hostname,
			"ip":       server.IP,
			"status":   status,
		})
	}

	a.runner.OnLog = func(serverIP, hostname, line string) {
		runtime.EventsEmit(a.ctx, "log", map[string]interface{}{
			"serverIP": serverIP,
			"hostname": hostname,
			"line":     line,
		})
	}

	a.runner.OnResult = func(result cisco.ExecutionResult) {
		logPath := strings.ReplaceAll(result.LogPath, "\\", "/")
		runtime.EventsEmit(a.ctx, "result", map[string]interface{}{
			"hostname": result.Server.Hostname,
			"ip":       result.Server.IP,
			"success":  result.Success,
			"error":    result.Error,
			"logPath":  logPath,
			"duration": result.Duration,
		})

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

	os.MkdirAll(logsDir, 0755)

	absPath, err := filepath.Abs(logsDir)
	if err != nil {
		absPath = logsDir
	}

	exec.Command("explorer.exe", absPath).Start()
}

// GetLogFiles returns list of log files in the logs directory
func (a *App) GetLogFiles() []map[string]string {
	logsDir := "logs"
	var files []map[string]string

	filepath.Walk(logsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".log" {
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

// GetCurrentVersion returns the current application version
func (a *App) GetCurrentVersion() string {
	return a.updater.GetCurrentVersion()
}

// CheckForUpdates checks for available updates
func (a *App) CheckForUpdates() *updater.UpdateInfo {
	info, err := a.updater.CheckForUpdates()
	if err != nil {
		runtime.EventsEmit(a.ctx, "updateError", err.Error())
		return nil
	}
	return info
}

// DownloadAndInstallUpdate downloads and installs the update
func (a *App) DownloadAndInstallUpdate(downloadURL string) bool {
	tempFile, err := a.updater.DownloadUpdate(downloadURL, func(downloaded, total int64) {
		percent := float64(downloaded) / float64(total) * 100
		runtime.EventsEmit(a.ctx, "updateProgress", map[string]interface{}{
			"downloaded": downloaded,
			"total":      total,
			"percent":    percent,
		})
	})

	if err != nil {
		runtime.EventsEmit(a.ctx, "updateError", "Download failed: "+err.Error())
		return false
	}

	if err := a.updater.ApplyUpdate(tempFile); err != nil {
		runtime.EventsEmit(a.ctx, "updateError", "Install failed: "+err.Error())
		return false
	}

	runtime.EventsEmit(a.ctx, "updateComplete", "Update installed successfully. Please restart the application.")
	return true
}

// RestartApp restarts the application
func (a *App) RestartApp() {
	exePath, err := os.Executable()
	if err != nil {
		runtime.EventsEmit(a.ctx, "updateError", "Failed to restart: "+err.Error())
		return
	}

	cmd := exec.Command(exePath)
	cmd.Start()
	os.Exit(0)
}

// ExportResults exports execution results to Excel file
func (a *App) ExportResults() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.runner == nil {
		return ""
	}

	results := a.runner.GetResults()
	if len(results) == 0 {
		return ""
	}

	outputPath := filepath.Join(a.runner.LogDir, "results.xlsx")

	err := cisco.ExportToExcel(results, a.commands, outputPath)
	if err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to export Excel: "+err.Error())
		return ""
	}

	return outputPath
}
