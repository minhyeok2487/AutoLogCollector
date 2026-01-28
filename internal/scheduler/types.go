package scheduler

import (
	"time"

	"cisco-plink/internal/cisco"
)

// ScheduledTask represents a scheduled execution task
type ScheduledTask struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`

	// Schedule configuration
	ScheduleType string `json:"scheduleType"` // "daily", "weekly", "monthly"
	Time         string `json:"time"`         // "HH:MM" (24-hour format)
	DaysOfWeek   []int  `json:"daysOfWeek"`   // 0-6 (Sunday-Saturday) for weekly
	DayOfMonth   int    `json:"dayOfMonth"`   // 1-31 for monthly

	// Credentials
	Username       string `json:"username"`
	Password       string `json:"password"`
	EnablePassword string `json:"enablePassword,omitempty"`

	// Execution configuration
	Servers         []cisco.Server `json:"servers"`
	Commands        []string       `json:"commands"`
	Timeout         int            `json:"timeout"`
	EnableMode      bool           `json:"enableMode"`
	DisablePaging   bool           `json:"disablePaging"`
	AutoExportExcel bool           `json:"autoExportExcel"`

	// Metadata
	LastRun *time.Time `json:"lastRun,omitempty"`
	NextRun *time.Time `json:"nextRun,omitempty"`
}

// ScheduleCallback is called when a scheduled task should be executed
type ScheduleCallback func(task *ScheduledTask)
