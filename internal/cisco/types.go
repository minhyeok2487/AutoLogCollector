package cisco

// Server represents a Cisco device
type Server struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
}

// Credentials holds SSH login information
type Credentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

// ExecutionResult represents the result of executing commands on a server
type ExecutionResult struct {
	Server   Server `json:"server"`
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	LogPath  string `json:"logPath,omitempty"`
	Duration int64  `json:"duration"` // milliseconds
}

// ProgressCallback is called when there's progress to report
type ProgressCallback func(current, total int, server Server, status string)

// ResultCallback is called when a server execution completes
type ResultCallback func(result ExecutionResult)
