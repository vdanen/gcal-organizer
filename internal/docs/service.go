// Package docs provides Google Docs operations.
package docs

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jflowers/gcal-organizer/pkg/models"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// ProcessedEmoji is the emoji used to mark processed action items.
const ProcessedEmoji = "🆔"

// AssigneeEmoji is the emoji used before the assignee name.
const AssigneeEmoji = "👤"

// DateEmoji is the emoji used before the date.
const DateEmoji = "📅"

// Service provides Google Docs operations.
type Service struct {
	client *docs.Service
}

// CheckboxItem represents a checkbox item found in a document.
type CheckboxItem struct {
	// Text is the text content of the checkbox item
	Text string

	// StartIndex is the position in the document
	StartIndex int64

	// EndIndex is the end position in the document
	EndIndex int64

	// IsChecked indicates if the checkbox is checked
	IsChecked bool

	// IsProcessed indicates if the item has already been processed (contains 🆔)
	IsProcessed bool
}

// NewService creates a new Docs service.
func NewService(ctx context.Context, httpClient *http.Client) (*Service, error) {
	srv, err := docs.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Docs service: %w", err)
	}

	return &Service{client: srv}, nil
}

// GetDocument retrieves a document by ID.
func (s *Service) GetDocument(ctx context.Context, docID string) (*docs.Document, error) {
	doc, err := s.client.Documents.Get(docID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return doc, nil
}

// ExtractCheckboxItems finds all checkbox list items in a document.
func (s *Service) ExtractCheckboxItems(ctx context.Context, docID string) ([]*CheckboxItem, error) {
	doc, err := s.GetDocument(ctx, docID)
	if err != nil {
		return nil, err
	}

	var items []*CheckboxItem

	// Iterate through document content
	for _, elem := range doc.Body.Content {
		if elem.Paragraph == nil {
			continue
		}

		para := elem.Paragraph

		// Check if this is a list item with checkbox
		if para.Bullet == nil {
			continue
		}

		// Extract text from paragraph
		var text strings.Builder
		for _, textElem := range para.Elements {
			if textElem.TextRun != nil {
				text.WriteString(textElem.TextRun.Content)
			}
		}

		content := strings.TrimSpace(text.String())
		if content == "" {
			continue
		}

		// Check if already processed
		isProcessed := strings.Contains(content, ProcessedEmoji)

		// Determine if it's a checkbox by checking the list properties
		// In Google Docs API, checkboxes are represented as list items
		// We'll consider any list item as a potential action item
		isChecked := false // Would need to check nesting level for actual checkbox state

		items = append(items, &CheckboxItem{
			Text:        content,
			StartIndex:  elem.StartIndex,
			EndIndex:    elem.EndIndex,
			IsChecked:   isChecked,
			IsProcessed: isProcessed,
		})
	}

	return items, nil
}

// AnnotateActionItem adds the processed marker to an action item in the document.
func (s *Service) AnnotateActionItem(ctx context.Context, docID string, item *models.ActionItem) error {
	annotation := fmt.Sprintf(" %s %s %s %s %s",
		AssigneeEmoji, item.Assignee,
		DateEmoji, item.DueDate.Format("2006-01-02"),
		ProcessedEmoji)

	// Create the request to insert text at the end of the line
	req := &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{
						Index: int64(item.LineIndex),
					},
					Text: annotation,
				},
			},
		},
	}

	_, err := s.client.Documents.BatchUpdate(docID, req).Do()
	if err != nil {
		return fmt.Errorf("failed to annotate document: %w", err)
	}

	return nil
}
