package tasks

import (
	"testing"

	"github.com/jflowers/gcal-organizer/pkg/models"
)

func TestFormatTaskTitle(t *testing.T) {
	tests := []struct {
		name     string
		docName  string
		taskText string
		want     string
	}{
		{
			name:     "short text",
			docName:  "Team Standup",
			taskText: "Review budget",
			want:     "[Team Standup] Review budget",
		},
		{
			name:     "long doc name truncated",
			docName:  "Weekly Engineering Team Standup Meeting Notes",
			taskText: "Review budget",
			want:     "[Weekly Engineering Team Sta...] Review budget",
		},
		{
			name:     "long task text truncated",
			docName:  "Standup",
			taskText: "This is a very long task description that should be truncated because it exceeds the maximum allowed length for task titles in the system",
			want:     "[Standup] This is a very long task description that should be truncated because it exceeds the maximum allo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTaskTitle(tt.docName, tt.taskText)
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestActionItemModel(t *testing.T) {
	item := &models.ActionItem{
		DocumentID:   "doc123",
		DocumentName: "Test Meeting - 2026-02-01",
		Text:         "John to review the proposal by Friday",
		Assignee:     "John",
	}

	if item.Assignee != "John" {
		t.Errorf("expected assignee John, got %s", item.Assignee)
	}
	if item.IsProcessed {
		t.Error("expected IsProcessed to be false")
	}
}
