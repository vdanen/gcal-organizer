// Package tasks provides Google Tasks operations.
package tasks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jflowers/gcal-organizer/pkg/models"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

// Service provides Google Tasks operations.
type Service struct {
	client     *tasks.Service
	taskListID string
}

// NewService creates a new Tasks service.
func NewService(ctx context.Context, httpClient *http.Client) (*Service, error) {
	srv, err := tasks.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Tasks service: %w", err)
	}

	return &Service{client: srv}, nil
}

// SetDefaultTaskList sets the task list to use (defaults to primary).
func (s *Service) SetDefaultTaskList(ctx context.Context) error {
	lists, err := s.client.Tasklists.List().Do()
	if err != nil {
		return fmt.Errorf("failed to list task lists: %w", err)
	}

	if len(lists.Items) == 0 {
		return fmt.Errorf("no task lists found")
	}

	// Use the first (default) task list
	s.taskListID = lists.Items[0].Id
	return nil
}

// CreateTask creates a new task from an action item.
func (s *Service) CreateTask(ctx context.Context, item *models.ActionItem) (string, error) {
	if s.taskListID == "" {
		if err := s.SetDefaultTaskList(ctx); err != nil {
			return "", err
		}
	}

	title := formatTaskTitle(item.DocumentName, item.Text)

	task := &tasks.Task{
		Title: title,
	}

	// Set due date if available
	if !item.DueDate.IsZero() {
		task.Due = item.DueDate.Format("2006-01-02T00:00:00Z")
	}

	// Add notes with document link
	task.Notes = fmt.Sprintf("From: %s\nAssignee: %s", item.DocumentName, item.Assignee)

	created, err := s.client.Tasks.Insert(s.taskListID, task).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	return created.Id, nil
}

// formatTaskTitle creates a task title in the format "[DocName] Task text"
func formatTaskTitle(docName, taskText string) string {
	// Truncate task text if too long
	maxLen := 100
	if len(taskText) > maxLen {
		taskText = taskText[:maxLen-3] + "..."
	}

	// Clean up the doc name
	if len(docName) > 30 {
		docName = docName[:27] + "..."
	}

	return fmt.Sprintf("[%s] %s", docName, taskText)
}

// ListTasks lists tasks from the default task list.
func (s *Service) ListTasks(ctx context.Context) ([]*tasks.Task, error) {
	if s.taskListID == "" {
		if err := s.SetDefaultTaskList(ctx); err != nil {
			return nil, err
		}
	}

	result, err := s.client.Tasks.List(s.taskListID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return result.Items, nil
}
