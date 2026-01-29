package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// Scheduler manages scheduled tasks
type Scheduler struct {
	cron      *cron.Cron
	tasks     map[string]*ScheduledTask
	cronIDs   map[string]cron.EntryID
	mu        sync.RWMutex
	onExecute ScheduleCallback
}

// NewScheduler creates a new scheduler instance
func NewScheduler(onExecute ScheduleCallback) *Scheduler {
	return &Scheduler{
		cron:      cron.New(),
		tasks:     make(map[string]*ScheduledTask),
		cronIDs:   make(map[string]cron.EntryID),
		onExecute: onExecute,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// AddTask adds a new scheduled task
func (s *Scheduler) AddTask(task *ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	// Check for duplicate name
	for _, t := range s.tasks {
		if t.ID != task.ID && t.Name == task.Name {
			return fmt.Errorf("schedule name '%s' already exists", task.Name)
		}
	}

	s.tasks[task.ID] = task

	if task.Enabled {
		if err := s.scheduleTask(task); err != nil {
			return err
		}
	}

	return nil
}

// UpdateTask updates an existing task
func (s *Scheduler) UpdateTask(task *ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate name
	for _, t := range s.tasks {
		if t.ID != task.ID && t.Name == task.Name {
			return fmt.Errorf("schedule name '%s' already exists", task.Name)
		}
	}

	// Remove old cron job if exists
	if entryID, exists := s.cronIDs[task.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.cronIDs, task.ID)
	}

	s.tasks[task.ID] = task

	if task.Enabled {
		if err := s.scheduleTask(task); err != nil {
			return err
		}
	}

	return nil
}

// DeleteTask removes a task
func (s *Scheduler) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.cronIDs[id]; exists {
		s.cron.Remove(entryID)
		delete(s.cronIDs, id)
	}

	delete(s.tasks, id)
	return nil
}

// GetTask returns a task by ID
func (s *Scheduler) GetTask(id string) *ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[id]
}

// GetTasks returns all tasks
func (s *Scheduler) GetTasks() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		// Update next run time
		if entryID, exists := s.cronIDs[task.ID]; exists {
			entry := s.cron.Entry(entryID)
			if !entry.Next.IsZero() {
				next := entry.Next
				task.NextRun = &next
			}
		}
		tasks = append(tasks, task)
	}
	return tasks
}

// ToggleTask enables or disables a task
func (s *Scheduler) ToggleTask(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	task.Enabled = enabled

	// Remove existing cron job
	if entryID, exists := s.cronIDs[id]; exists {
		s.cron.Remove(entryID)
		delete(s.cronIDs, id)
	}

	// Re-schedule if enabled
	if enabled {
		if err := s.scheduleTask(task); err != nil {
			return err
		}
	}

	return nil
}

// LoadTasks loads tasks from a list (called on startup)
func (s *Scheduler) LoadTasks(tasks []*ScheduledTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range tasks {
		s.tasks[task.ID] = task
		if task.Enabled {
			s.scheduleTask(task)
		}
	}
}

// scheduleTask adds a cron job for the task (must be called with lock held)
func (s *Scheduler) scheduleTask(task *ScheduledTask) error {
	cronExpr, err := s.buildCronExpression(task)
	if err != nil {
		return err
	}

	taskID := task.ID
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeTask(taskID)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule task: %v", err)
	}

	s.cronIDs[task.ID] = entryID

	// Update next run time
	entry := s.cron.Entry(entryID)
	if !entry.Next.IsZero() {
		next := entry.Next
		task.NextRun = &next
	}

	return nil
}

// buildCronExpression creates a cron expression from task schedule
func (s *Scheduler) buildCronExpression(task *ScheduledTask) (string, error) {
	parts := strings.Split(task.Time, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid time format: %s", task.Time)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return "", fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return "", fmt.Errorf("invalid minute: %s", parts[1])
	}

	switch task.ScheduleType {
	case "daily":
		// Run every day at specified time
		return fmt.Sprintf("%d %d * * *", minute, hour), nil

	case "weekly":
		// Run on specified days of week
		if len(task.DaysOfWeek) == 0 {
			return "", fmt.Errorf("no days specified for weekly schedule")
		}
		days := make([]string, len(task.DaysOfWeek))
		for i, d := range task.DaysOfWeek {
			days[i] = strconv.Itoa(d)
		}
		return fmt.Sprintf("%d %d * * %s", minute, hour, strings.Join(days, ",")), nil

	case "monthly":
		// Run on specified day of month
		if task.DayOfMonth < 1 || task.DayOfMonth > 31 {
			return "", fmt.Errorf("invalid day of month: %d", task.DayOfMonth)
		}
		return fmt.Sprintf("%d %d %d * *", minute, hour, task.DayOfMonth), nil

	default:
		return "", fmt.Errorf("unknown schedule type: %s", task.ScheduleType)
	}
}

// executeTask runs a scheduled task
func (s *Scheduler) executeTask(taskID string) {
	s.mu.Lock()
	task, exists := s.tasks[taskID]
	if !exists || !task.Enabled {
		s.mu.Unlock()
		return
	}

	// Update last run time
	now := time.Now()
	task.LastRun = &now
	s.mu.Unlock()

	// Call the execution callback
	if s.onExecute != nil {
		s.onExecute(task)
	}
}
