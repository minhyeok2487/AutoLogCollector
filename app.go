package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"cisco-plink/internal/cisco"
	"cisco-plink/internal/config"
	appCrypto "cisco-plink/internal/crypto"
	"cisco-plink/internal/scheduler"
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
	ctx             context.Context
	runner          *cisco.Runner
	updater         *updater.Updater
	scheduler       *scheduler.Scheduler
	mu              sync.Mutex
	servers         []cisco.Server
	commands        []string
	autoExportExcel bool // Flag for auto Excel export on completion
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.updater = updater.NewUpdater(GitHubOwner, GitHubRepo)

	// Initialize scheduler
	a.scheduler = scheduler.NewScheduler(a.executeScheduledTask)
	a.scheduler.Start()

	// Load saved schedules
	cfg, err := config.Load()
	if err == nil && len(cfg.Schedules) > 0 {
		a.scheduler.LoadTasks(cfg.Schedules)
	}
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
}

// executeScheduledTask is called when a scheduled task triggers
func (a *App) executeScheduledTask(task *scheduler.ScheduledTask) {
	if task.Username == "" || task.Password == "" {
		runtime.EventsEmit(a.ctx, "scheduleSkipped", map[string]interface{}{
			"taskId":   task.ID,
			"taskName": task.Name,
			"reason":   "No credentials configured in schedule.",
		})
		return
	}

	// Check if another execution is running
	if a.runner != nil && a.runner.IsRunning() {
		runtime.EventsEmit(a.ctx, "scheduleSkipped", map[string]interface{}{
			"taskId":   task.ID,
			"taskName": task.Name,
			"reason":   "Another execution is already running.",
		})
		return
	}

	// Emit schedule started event
	runtime.EventsEmit(a.ctx, "scheduleStarted", map[string]interface{}{
		"taskId":   task.ID,
		"taskName": task.Name,
	})

	// Set servers and commands from task
	a.mu.Lock()
	a.servers = task.Servers
	a.commands = task.Commands
	a.mu.Unlock()

	// Start execution with task's credentials and options
	a.StartExecution(task.Username, task.Password, task.Timeout, task.EnableMode, task.DisablePaging, task.AutoExportExcel, task.EnablePassword)
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
				IP:             ip,
				Hostname:       hostname,
				Username:       s["username"],
				Password:       s["password"],
				EnablePassword: s["enablePassword"],
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

// SaveServerList saves the current server list to config/servers.json (encrypted)
func (a *App) SaveServerList(servers []map[string]string) bool {
	key, err := appCrypto.LoadOrGenerateKey()
	if err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to load encryption key: "+err.Error())
		return false
	}

	serverList := make([]cisco.Server, 0, len(servers))
	for _, s := range servers {
		ip := s["ip"]
		if ip == "" {
			continue
		}
		hostname := s["hostname"]
		if hostname == "" {
			hostname = ip
		}
		srv := cisco.Server{
			IP:       ip,
			Hostname: hostname,
		}
		// Encrypt per-server credentials
		if s["username"] != "" {
			srv.Username = s["username"]
			encPwd, encEnPwd, err := appCrypto.EncryptFields(s["password"], s["enablePassword"], key)
			if err == nil {
				srv.Password = encPwd
				srv.EnablePassword = encEnPwd
			}
		}
		serverList = append(serverList, srv)
	}

	data, err := json.MarshalIndent(serverList, "", "  ")
	if err != nil {
		return false
	}

	if err := os.MkdirAll("config", 0755); err != nil {
		return false
	}
	return os.WriteFile(filepath.Join("config", "servers.json"), data, 0644) == nil
}

// LoadServerList loads saved server list from config/servers.json (decrypted)
func (a *App) LoadServerList() []map[string]string {
	data, err := os.ReadFile(filepath.Join("config", "servers.json"))
	if err != nil {
		return nil
	}

	var servers []cisco.Server
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil
	}

	key, _ := appCrypto.LoadOrGenerateKey()

	result := make([]map[string]string, len(servers))
	for i, s := range servers {
		// Decrypt per-server credentials
		if key != nil && s.Password != "" {
			s.Password, s.EnablePassword = appCrypto.DecryptFields(s.Password, s.EnablePassword, key)
		}
		result[i] = map[string]string{
			"ip":             s.IP,
			"hostname":       s.Hostname,
			"username":       s.Username,
			"password":       s.Password,
			"enablePassword": s.EnablePassword,
		}
	}
	return result
}

// ExportServersToCSV exports server list to a CSV file
func (a *App) ExportServersToCSV(servers []map[string]string) bool {
	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Servers to CSV",
		DefaultFilename: "servers.csv",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
		},
	})
	if err != nil || file == "" {
		return false
	}

	var lines []string
	lines = append(lines, "ip,hostname")
	for _, s := range servers {
		ip := s["ip"]
		hostname := s["hostname"]
		if ip != "" {
			lines = append(lines, ip+","+hostname)
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(file, []byte(content), 0644) == nil
}

// ImportCommandsFromTxt opens file dialog and returns commands as text
func (a *App) ImportCommandsFromTxt() string {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import Commands from TXT",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil || file == "" {
		return ""
	}

	data, err := os.ReadFile(file)
	if err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to read file: "+err.Error())
		return ""
	}
	return string(data)
}

// ExportCommandsToTxt saves commands to a TXT file
func (a *App) ExportCommandsToTxt(commands string) bool {
	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Commands to TXT",
		DefaultFilename: "commands.txt",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
		},
	})
	if err != nil || file == "" {
		return false
	}
	return os.WriteFile(file, []byte(commands), 0644) == nil
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
func (a *App) StartExecution(username, password string, timeout int, enableMode, disablePaging, autoExportExcel bool, enablePassword string) bool {
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

	// Store autoExportExcel flag for completion event
	a.autoExportExcel = autoExportExcel

	creds := &cisco.Credentials{
		User:           username,
		Password:       password,
		EnablePassword: enablePassword,
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
				"success":         success,
				"fail":            fail,
				"total":           total,
				"logDir":          a.runner.LogDir,
				"autoExportExcel": a.autoExportExcel,
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

// ==================== Schedule Management ====================

// CreateSchedule creates a new scheduled task
func (a *App) CreateSchedule(taskData map[string]interface{}) string {
	task := a.mapToScheduledTask(taskData)
	task.Enabled = true

	if err := a.scheduler.AddTask(task); err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to create schedule: "+err.Error())
		return ""
	}

	a.saveSchedules()
	return task.ID
}

// UpdateSchedule updates an existing scheduled task
func (a *App) UpdateSchedule(taskData map[string]interface{}) bool {
	task := a.mapToScheduledTask(taskData)

	if err := a.scheduler.UpdateTask(task); err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to update schedule: "+err.Error())
		return false
	}

	a.saveSchedules()
	return true
}

// DeleteSchedule removes a scheduled task
func (a *App) DeleteSchedule(id string) bool {
	if err := a.scheduler.DeleteTask(id); err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to delete schedule: "+err.Error())
		return false
	}

	a.saveSchedules()
	return true
}

// GetSchedules returns all scheduled tasks
func (a *App) GetSchedules() []map[string]interface{} {
	tasks := a.scheduler.GetTasks()
	result := make([]map[string]interface{}, len(tasks))

	for i, task := range tasks {
		result[i] = a.scheduledTaskToMap(task)
	}

	return result
}

// ToggleSchedule enables or disables a scheduled task
func (a *App) ToggleSchedule(id string, enabled bool) bool {
	if err := a.scheduler.ToggleTask(id, enabled); err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to toggle schedule: "+err.Error())
		return false
	}

	a.saveSchedules()
	return true
}

// RunScheduleNow manually triggers a scheduled task
func (a *App) RunScheduleNow(id string) bool {
	task := a.scheduler.GetTask(id)
	if task == nil {
		runtime.EventsEmit(a.ctx, "error", "Schedule not found")
		return false
	}

	go a.executeScheduledTask(task)
	return true
}

// saveSchedules saves all schedules to config file
func (a *App) saveSchedules() {
	tasks := a.scheduler.GetTasks()
	if err := config.SaveSchedules(tasks); err != nil {
		runtime.EventsEmit(a.ctx, "error", "Failed to save schedules: "+err.Error())
	}
}

// mapToScheduledTask converts a map to ScheduledTask
func (a *App) mapToScheduledTask(data map[string]interface{}) *scheduler.ScheduledTask {
	task := &scheduler.ScheduledTask{
		Timeout:       1,
		DisablePaging: true,
	}

	if id, ok := data["id"].(string); ok {
		task.ID = id
	}
	if name, ok := data["name"].(string); ok {
		task.Name = name
	}
	if enabled, ok := data["enabled"].(bool); ok {
		task.Enabled = enabled
	}
	if scheduleType, ok := data["scheduleType"].(string); ok {
		task.ScheduleType = scheduleType
	}
	if timeStr, ok := data["time"].(string); ok {
		task.Time = timeStr
	}
	if daysOfWeek, ok := data["daysOfWeek"].([]interface{}); ok {
		task.DaysOfWeek = make([]int, len(daysOfWeek))
		for i, d := range daysOfWeek {
			if day, ok := d.(float64); ok {
				task.DaysOfWeek[i] = int(day)
			}
		}
	}
	if dayOfMonth, ok := data["dayOfMonth"].(float64); ok {
		task.DayOfMonth = int(dayOfMonth)
	}
	if timeout, ok := data["timeout"].(float64); ok {
		task.Timeout = int(timeout)
	}
	if enableMode, ok := data["enableMode"].(bool); ok {
		task.EnableMode = enableMode
	}
	if disablePaging, ok := data["disablePaging"].(bool); ok {
		task.DisablePaging = disablePaging
	}
	if autoExportExcel, ok := data["autoExportExcel"].(bool); ok {
		task.AutoExportExcel = autoExportExcel
	}

	// Parse credentials
	if username, ok := data["username"].(string); ok {
		task.Username = username
	}
	if password, ok := data["password"].(string); ok {
		task.Password = password
	}
	if enablePassword, ok := data["enablePassword"].(string); ok {
		task.EnablePassword = enablePassword
	}

	// Parse servers
	if servers, ok := data["servers"].([]interface{}); ok {
		task.Servers = make([]cisco.Server, 0, len(servers))
		for _, s := range servers {
			if serverMap, ok := s.(map[string]interface{}); ok {
				server := cisco.Server{}
				if ip, ok := serverMap["ip"].(string); ok {
					server.IP = ip
				}
				if hostname, ok := serverMap["hostname"].(string); ok {
					server.Hostname = hostname
				}
				if username, ok := serverMap["username"].(string); ok {
					server.Username = username
				}
				if password, ok := serverMap["password"].(string); ok {
					server.Password = password
				}
				if enablePassword, ok := serverMap["enablePassword"].(string); ok {
					server.EnablePassword = enablePassword
				}
				if server.IP != "" {
					if server.Hostname == "" {
						server.Hostname = server.IP
					}
					task.Servers = append(task.Servers, server)
				}
			}
		}
	}

	// Parse commands
	if commands, ok := data["commands"].([]interface{}); ok {
		task.Commands = make([]string, 0, len(commands))
		for _, c := range commands {
			if cmd, ok := c.(string); ok && cmd != "" {
				task.Commands = append(task.Commands, cmd)
			}
		}
	}

	return task
}

// scheduledTaskToMap converts a ScheduledTask to map
func (a *App) scheduledTaskToMap(task *scheduler.ScheduledTask) map[string]interface{} {
	servers := make([]map[string]string, len(task.Servers))
	for i, s := range task.Servers {
		servers[i] = map[string]string{
			"ip":             s.IP,
			"hostname":       s.Hostname,
			"username":       s.Username,
			"password":       s.Password,
			"enablePassword": s.EnablePassword,
		}
	}

	result := map[string]interface{}{
		"id":              task.ID,
		"name":            task.Name,
		"enabled":         task.Enabled,
		"scheduleType":    task.ScheduleType,
		"time":            task.Time,
		"daysOfWeek":      task.DaysOfWeek,
		"dayOfMonth":      task.DayOfMonth,
		"username":        task.Username,
		"password":        task.Password,
		"enablePassword":  task.EnablePassword,
		"servers":         servers,
		"commands":        task.Commands,
		"timeout":         task.Timeout,
		"enableMode":      task.EnableMode,
		"disablePaging":   task.DisablePaging,
		"autoExportExcel": task.AutoExportExcel,
	}

	if task.LastRun != nil {
		result["lastRun"] = task.LastRun.Format(time.RFC3339)
	}
	if task.NextRun != nil {
		result["nextRun"] = task.NextRun.Format(time.RFC3339)
	}

	return result
}
