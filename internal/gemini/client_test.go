package gemini

import (
	"testing"
)

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		wantAssignee string
		wantDate     string
		wantError    bool
	}{
		{
			name:         "valid response",
			response:     `{"assignee": "John", "date": "2026-02-07"}`,
			wantAssignee: "John",
			wantDate:     "2026-02-07",
			wantError:    false,
		},
		{
			name:         "with markdown code block",
			response:     "```json\n{\"assignee\": \"Sarah\", \"date\": \"2026-02-10\"}\n```",
			wantAssignee: "Sarah",
			wantDate:     "2026-02-10",
			wantError:    false,
		},
		{
			name:         "null values",
			response:     `{"assignee": null, "date": null}`,
			wantAssignee: "",
			wantDate:     "",
			wantError:    false,
		},
		{
			name:         "only assignee",
			response:     `{"assignee": "Mike", "date": null}`,
			wantAssignee: "Mike",
			wantDate:     "",
			wantError:    false,
		},
		{
			name:         "with extra whitespace",
			response:     "  \n  {\"assignee\": \"Alex\", \"date\": \"2026-02-15\"}  \n  ",
			wantAssignee: "Alex",
			wantDate:     "2026-02-15",
			wantError:    false,
		},
		{
			name:      "invalid json",
			response:  "not json at all",
			wantError: true,
		},
		{
			name:         "embedded in text",
			response:     "Here is the result: {\"assignee\": \"Bob\", \"date\": \"2026-03-01\"} as requested.",
			wantAssignee: "Bob",
			wantDate:     "2026-03-01",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseResponse(tt.response)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Assignee != tt.wantAssignee {
				t.Errorf("assignee: got %q, want %q", result.Assignee, tt.wantAssignee)
			}
			if result.Date != tt.wantDate {
				t.Errorf("date: got %q, want %q", result.Date, tt.wantDate)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name      string
		dateStr   string
		wantError bool
	}{
		{"valid date", "2026-02-07", false},
		{"empty string", "", true},
		{"invalid format", "02-07-2026", true},
		{"partial date", "2026-02", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDate(tt.dateStr)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
