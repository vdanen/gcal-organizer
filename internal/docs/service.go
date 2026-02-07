// Package docs provides Google Docs operations.
package docs

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"unicode"

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

// SuggestedNextStepsHeading is the section heading to look for.
const SuggestedNextStepsHeading = "Suggested next steps"

// NotesTabName is the tab name to target.
const NotesTabName = "Notes"

// ExtractCheckboxItems finds checkbox items in the "Suggested next steps" section of the "Notes" tab.
func (s *Service) ExtractCheckboxItems(ctx context.Context, docID string) ([]*CheckboxItem, error) {
	doc, err := s.GetDocument(ctx, docID)
	if err != nil {
		return nil, err
	}

	// Find the content to process - either from the "Notes" tab or the main body
	var content []*docs.StructuralElement

	if len(doc.Tabs) > 0 {
		// Multi-tab document - look for "Notes" tab
		for _, tab := range doc.Tabs {
			if tab.TabProperties != nil && tab.TabProperties.Title == NotesTabName {
				if tab.DocumentTab != nil && tab.DocumentTab.Body != nil {
					content = tab.DocumentTab.Body.Content
				}
				break
			}
		}
		if content == nil {
			// Notes tab not found
			return nil, nil
		}
	} else if doc.Body != nil {
		// Single-tab or legacy document - use main body
		content = doc.Body.Content
	} else {
		return nil, nil
	}

	return s.extractItemsFromSection(content)
}

// extractItemsFromSection extracts checkbox items from the "Suggested next steps" section.
func (s *Service) extractItemsFromSection(content []*docs.StructuralElement) ([]*CheckboxItem, error) {
	var items []*CheckboxItem
	inSuggestedSection := false

	for _, elem := range content {
		if elem.Paragraph == nil {
			continue
		}

		para := elem.Paragraph

		// Check if this is the "Suggested next steps" heading
		paraText := extractParagraphText(para)
		if strings.Contains(strings.ToLower(paraText), strings.ToLower(SuggestedNextStepsHeading)) {
			inSuggestedSection = true
			continue
		}

		// If we haven't found the section yet, skip
		if !inSuggestedSection {
			continue
		}

		// Check if this is a list item (bullet or checkbox)
		if para.Bullet == nil {
			continue
		}

		content := strings.TrimSpace(paraText)
		if content == "" {
			continue
		}

		// Check if already processed
		isProcessed := strings.Contains(content, ProcessedEmoji)

		items = append(items, &CheckboxItem{
			Text:        content,
			StartIndex:  elem.StartIndex,
			EndIndex:    elem.EndIndex,
			IsChecked:   false,
			IsProcessed: isProcessed,
		})
	}

	return items, nil
}

// extractParagraphText extracts all text from a paragraph.
func extractParagraphText(para *docs.Paragraph) string {
	var text strings.Builder
	for _, textElem := range para.Elements {
		if textElem.TextRun != nil {
			text.WriteString(textElem.TextRun.Content)
		}
	}
	// Strip non-printable characters (vertical tabs, carriage returns, zero-width spaces, etc.)
	// that Google Docs API may include in TextRun.Content
	cleaned := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r == ' ' {
			return r // keep normal whitespace
		}
		if unicode.IsControl(r) || !unicode.IsPrint(r) {
			return -1 // strip
		}
		return r
	}, text.String())
	return strings.TrimSpace(cleaned)
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
