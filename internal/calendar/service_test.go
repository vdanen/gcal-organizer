package calendar

import (
	"testing"
)

func TestExtractDriveLinks(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name      string
		text      string
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "single doc link",
			text:      "Check out this doc: https://docs.google.com/document/d/1abc123XYZ/edit",
			wantCount: 1,
			wantIDs:   []string{"1abc123XYZ"},
		},
		{
			name:      "multiple links",
			text:      "Doc: https://docs.google.com/document/d/doc1 and Sheet: https://docs.google.com/spreadsheets/d/sheet2",
			wantCount: 2,
			wantIDs:   []string{"doc1", "sheet2"},
		},
		{
			name:      "drive file link",
			text:      "File: https://drive.google.com/file/d/fileID123/view",
			wantCount: 1,
			wantIDs:   []string{"fileID123"},
		},
		{
			name:      "presentation link",
			text:      "Slides: https://docs.google.com/presentation/d/slides456",
			wantCount: 1,
			wantIDs:   []string{"slides456"},
		},
		{
			name:      "no links",
			text:      "Just some regular text without any links",
			wantCount: 0,
			wantIDs:   nil,
		},
		{
			name:      "non-drive link",
			text:      "Check https://example.com/document/d/notadoc",
			wantCount: 0,
			wantIDs:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachments := s.extractDriveLinks(tt.text)

			if len(attachments) != tt.wantCount {
				t.Fatalf("got %d attachments, want %d", len(attachments), tt.wantCount)
			}

			for i, wantID := range tt.wantIDs {
				if attachments[i].FileID != wantID {
					t.Errorf("attachment[%d].FileID = %q, want %q", i, attachments[i].FileID, wantID)
				}
			}
		})
	}
}
