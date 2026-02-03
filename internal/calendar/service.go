// Package calendar provides Google Calendar operations.
package calendar

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/jflowers/gcal-organizer/pkg/models"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// DriveURLPattern matches Google Drive file URLs in text.
var DriveURLPattern = regexp.MustCompile(`https://(?:docs|drive)\.google\.com/(?:document|file|spreadsheets|presentation)/d/([a-zA-Z0-9_-]+)`)

// Service provides Google Calendar operations.
type Service struct {
	client *calendar.Service
}

// NewService creates a new Calendar service.
func NewService(ctx context.Context, httpClient *http.Client) (*Service, error) {
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Calendar service: %w", err)
	}

	return &Service{client: srv}, nil
}

// ListRecentEvents retrieves calendar events from the past N days.
func (s *Service) ListRecentEvents(ctx context.Context, daysBack int) ([]*models.CalendarEvent, error) {
	now := time.Now()
	minTime := now.AddDate(0, 0, -daysBack).Format(time.RFC3339)
	maxTime := now.Format(time.RFC3339)

	var events []*models.CalendarEvent
	pageToken := ""

	for {
		req := s.client.Events.List("primary").
			TimeMin(minTime).
			TimeMax(maxTime).
			SingleEvents(true).
			OrderBy("startTime")

		if pageToken != "" {
			req.PageToken(pageToken)
		}

		result, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list calendar events: %w", err)
		}

		for _, item := range result.Items {
			event, err := s.parseEvent(item)
			if err != nil {
				continue // Skip events we can't parse
			}
			events = append(events, event)
		}

		pageToken = result.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return events, nil
}

// parseEvent converts a calendar event to our model.
func (s *Service) parseEvent(item *calendar.Event) (*models.CalendarEvent, error) {
	var start, end time.Time

	if item.Start.DateTime != "" {
		var err error
		start, err = time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}
	} else if item.Start.Date != "" {
		var err error
		start, err = time.Parse("2006-01-02", item.Start.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start date: %w", err)
		}
	}

	if item.End.DateTime != "" {
		var err error
		end, err = time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}
	} else if item.End.Date != "" {
		var err error
		end, err = time.Parse("2006-01-02", item.End.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end date: %w", err)
		}
	}

	// Extract attachments
	var attachments []models.Attachment
	for _, att := range item.Attachments {
		attachments = append(attachments, models.Attachment{
			FileID:   att.FileId,
			Title:    att.Title,
			MimeType: att.MimeType,
			FileURL:  att.FileUrl,
		})
	}

	// Also extract Drive links from description
	descAttachments := s.extractDriveLinks(item.Description)
	attachments = append(attachments, descAttachments...)

	return &models.CalendarEvent{
		ID:          item.Id,
		Title:       item.Summary,
		Start:       start,
		End:         end,
		Description: item.Description,
		Attachments: attachments,
	}, nil
}

// extractDriveLinks finds Google Drive URLs in text.
func (s *Service) extractDriveLinks(text string) []models.Attachment {
	var attachments []models.Attachment

	matches := DriveURLPattern.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			attachments = append(attachments, models.Attachment{
				FileID:  match[1],
				FileURL: match[0],
			})
		}
	}

	return attachments
}
