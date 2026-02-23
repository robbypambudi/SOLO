package types

import (
	"context"
	"encoding/json"
	"time"
)

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Task represents a unit of work in the reasoning pipeline.
// Lightweight alternative to Ray's remote function/actor model.
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      TaskStatus             `json:"status"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Dependencies []string              `json:"dependencies,omitempty"` // Task IDs that must complete first
}

// TaskExecutor defines the interface for executing a task.
// Implementations can wrap LLM calls, retrievers, planners, etc.
type TaskExecutor interface {
	Execute(ctx context.Context, task *Task) (*Task, error)
}

// TaskFunc is a function that can execute a task.
type TaskFunc func(ctx context.Context, task *Task) (*Task, error)

// Execute implements TaskExecutor for TaskFunc.
func (f TaskFunc) Execute(ctx context.Context, task *Task) (*Task, error) {
	return f(ctx, task)
}

// MarshalPayload serializes the task payload to JSON.
func (t *Task) MarshalPayload() ([]byte, error) {
	return json.Marshal(t.Payload)
}
