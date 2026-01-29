package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"cisco-plink/internal/crypto"
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

// Load loads configuration from disk and decrypts sensitive fields
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

	// Decrypt passwords
	key, err := crypto.LoadOrGenerateKey()
	if err != nil {
		return cfg, nil // Return config with encrypted passwords if key fails
	}

	for _, task := range cfg.Schedules {
		task.Password, task.EnablePassword = crypto.DecryptFields(task.Password, task.EnablePassword, key)
	}

	return cfg, nil
}

// Save saves configuration to disk with sensitive fields encrypted
func Save(cfg *Config) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	key, err := crypto.LoadOrGenerateKey()
	if err != nil {
		return err
	}

	// Create a copy with encrypted passwords (don't modify in-memory tasks)
	saveCfg := &Config{
		Schedules: make([]*scheduler.ScheduledTask, len(cfg.Schedules)),
	}

	for i, task := range cfg.Schedules {
		taskCopy := *task
		encPwd, encEnPwd, err := crypto.EncryptFields(task.Password, task.EnablePassword, key)
		if err != nil {
			return err
		}
		taskCopy.Password = encPwd
		taskCopy.EnablePassword = encEnPwd
		saveCfg.Schedules[i] = &taskCopy
	}

	data, err := json.MarshalIndent(saveCfg, "", "  ")
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
