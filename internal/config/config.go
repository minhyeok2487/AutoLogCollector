package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"cisco-plink/internal/scheduler"
)

const (
	configDir     = "config"
	schedulesFile = "schedules.json"
)

// Config holds all application configuration
type Config struct {
	Schedules []*scheduler.ScheduledTask `json:"schedules"`
}

// Load loads configuration from disk
func Load() (*Config, error) {
	cfg := &Config{
		Schedules: make([]*scheduler.ScheduledTask, 0),
	}

	path := filepath.Join(configDir, schedulesFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return empty config if file doesn't exist
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves configuration to disk
func Save(cfg *Config) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, schedulesFile)
	return os.WriteFile(path, data, 0644)
}

// SaveSchedules saves only the schedules
func SaveSchedules(schedules []*scheduler.ScheduledTask) error {
	cfg := &Config{
		Schedules: schedules,
	}
	return Save(cfg)
}
