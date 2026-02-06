// Package organizer provides the main orchestration logic.
package organizer

import (
	"context"
	"fmt"
	"strings"

	"github.com/jflowers/gcal-organizer/internal/calendar"
	"github.com/jflowers/gcal-organizer/internal/config"
	"github.com/jflowers/gcal-organizer/internal/docs"
	"github.com/jflowers/gcal-organizer/internal/drive"
	"github.com/jflowers/gcal-organizer/internal/gemini"
	"github.com/jflowers/gcal-organizer/internal/tasks"
	"github.com/jflowers/gcal-organizer/pkg/models"
)

// Stats tracks operation counts for summary reporting.
type Stats struct {
	DocumentsFound    int
	DocumentsMoved    int
	ShortcutsCreated  int
	ShortcutsTrashed  int
	EventsProcessed   int
	EventsWithAttach  int
	TasksCreated      int
	ItemsSkipped      int
	Errors            int
}

// Organizer orchestrates all the services.
type Organizer struct {
	config   *config.Config
	drive    *drive.Service
	docs     *docs.Service
	calendar *calendar.Service
	tasks    *tasks.Service
	gemini   *gemini.Client

	verbose bool
	stats   Stats
}

// New creates a new Organizer with all services initialized.
func New(ctx context.Context, cfg *config.Config, driveSvc *drive.Service, docsSvc *docs.Service, calSvc *calendar.Service, tasksSvc *tasks.Service, geminiClient *gemini.Client) *Organizer {
	return &Organizer{
		config:   cfg,
		drive:    driveSvc,
		docs:     docsSvc,
		calendar: calSvc,
		tasks:    tasksSvc,
		gemini:   geminiClient,
		verbose:  cfg.Verbose,
	}
}

// RunFullWorkflow executes all operations in sequence.
func (o *Organizer) RunFullWorkflow(ctx context.Context) error {
	if o.drive.IsDryRun() {
		o.log("═══════════════════════════════════════════════════════════")
		o.log("🔍 DRY RUN MODE - No changes will be made")
		o.log("═══════════════════════════════════════════════════════════")
	}
	o.log("🚀 Starting full workflow...")
	o.log("")

	// Step 1: Organize documents
	o.log("📁 STEP 1: Organizing Documents")
	o.log("───────────────────────────────────────────────────────────")
	if err := o.OrganizeDocuments(ctx); err != nil {
		return fmt.Errorf("organize documents failed: %w", err)
	}
	o.log("")

	// Step 2: Sync calendar
	o.log("📅 STEP 2: Syncing Calendar Attachments")
	o.log("───────────────────────────────────────────────────────────")
	if err := o.SyncCalendarAttachments(ctx); err != nil {
		return fmt.Errorf("sync calendar failed: %w", err)
	}
	o.log("")

	// Step 3: Extract and create tasks
	o.log("🤖 STEP 3: Extracting Action Items")
	o.log("───────────────────────────────────────────────────────────")
	if err := o.ExtractAndCreateTasks(ctx); err != nil {
		return fmt.Errorf("extract tasks failed: %w", err)
	}
	o.log("")

	// Print summary
	o.printSummary()

	return nil
}

// printSummary outputs the final statistics.
func (o *Organizer) printSummary() {
	o.log("═══════════════════════════════════════════════════════════")
	if o.drive.IsDryRun() {
		o.log("📊 DRY RUN SUMMARY - What would happen:")
	} else {
		o.log("📊 WORKFLOW SUMMARY:")
	}
	o.log("───────────────────────────────────────────────────────────")
	o.log(fmt.Sprintf("   📄 Documents found:        %d", o.stats.DocumentsFound))
	o.log(fmt.Sprintf("   📂 Documents moved:        %d", o.stats.DocumentsMoved))
	o.log(fmt.Sprintf("   🔗 Shortcuts created:      %d", o.stats.ShortcutsCreated))
	if o.stats.ShortcutsTrashed > 0 {
		o.log(fmt.Sprintf("   🗑️  Shortcuts trashed:      %d", o.stats.ShortcutsTrashed))
	}
	o.log(fmt.Sprintf("   📅 Events processed:       %d", o.stats.EventsProcessed))
	o.log(fmt.Sprintf("   📎 Events with attachments: %d", o.stats.EventsWithAttach))
	o.log(fmt.Sprintf("   ✅ Tasks created:          %d", o.stats.TasksCreated))
	o.log(fmt.Sprintf("   ⏭️  Items skipped:          %d", o.stats.ItemsSkipped))
	if o.stats.Errors > 0 {
		o.log(fmt.Sprintf("   ⚠️  Errors encountered:     %d", o.stats.Errors))
	}
	o.log("═══════════════════════════════════════════════════════════")
	if o.drive.IsDryRun() {
		o.log("✨ Dry run complete - no changes were made")
	} else {
		o.log("✅ Workflow complete!")
	}
}

// OrganizeDocuments finds meeting documents and organizes them into folders.
func (o *Organizer) OrganizeDocuments(ctx context.Context) error {
	// Set up master folder
	if err := o.drive.SetMasterFolder(ctx, o.config.MasterFolderName); err != nil {
		return err
	}

	// Find documents
	documents, err := o.drive.ListMeetingDocuments(ctx, o.config.FilenameKeywords)
	if err != nil {
		return err
	}

	o.stats.DocumentsFound = len(documents)
	o.log(fmt.Sprintf("   Found %d meeting documents", len(documents)))
	o.log("")

	for _, doc := range documents {
		// Get or create meeting folder
		folder, err := o.drive.GetOrCreateMeetingFolder(ctx, doc.MeetingName)
		if err != nil {
			o.log(fmt.Sprintf("   ⚠️  Error: Failed to get/create folder for %s: %v", doc.MeetingName, err))
			o.stats.Errors++
			continue
		}

		var result drive.ActionResult
		if doc.IsOwned {
			// For owned fallback files, also clean up any redundant shortcut
			if doc.IsFallback && folder.ID != "" {
				shortcutID, err := o.drive.FindShortcutToFile(ctx, doc.ID, folder.ID)
				if err != nil {
					o.logVerbose(fmt.Sprintf("   ⚠️  Could not check for shortcuts: %v", err))
				} else if shortcutID != "" {
					// Found a shortcut pointing to this file - trash it
					trashResult := o.drive.TrashFile(ctx, shortcutID, 
						fmt.Sprintf("Trash redundant shortcut to %s (file being moved)", doc.Name))
					if !trashResult.Skipped || trashResult.Reason == "dry-run" {
						o.stats.ShortcutsTrashed++
						o.log(fmt.Sprintf("   🗑️  %s", trashResult.Details))
					}
				}
			}
			result = o.drive.MoveDocument(ctx, doc.ID, doc.Name, doc.ParentFolderID, folder.ID, folder.Name)
		} else {
			result = o.drive.CreateShortcut(ctx, doc.ID, doc.Name, folder.ID, folder.Name, folder.IsNew)
		}

		o.logActionResult(result, doc.IsOwned)
	}

	return nil
}

// SyncCalendarAttachments syncs calendar event attachments to meeting folders.
func (o *Organizer) SyncCalendarAttachments(ctx context.Context) error {
	events, err := o.calendar.ListRecentEvents(ctx, o.config.DaysToLookBack)
	if err != nil {
		return err
	}

	o.stats.EventsProcessed = len(events)
	o.log(fmt.Sprintf("   Found %d calendar events (last %d days)", len(events), o.config.DaysToLookBack))
	o.log("")

	for _, event := range events {
		if len(event.Attachments) == 0 {
			continue
		}

		o.stats.EventsWithAttach++

		// Get or create meeting folder
		folder, err := o.drive.GetOrCreateMeetingFolder(ctx, event.Title)
		if err != nil {
			o.log(fmt.Sprintf("   ⚠️  Error: Failed to get/create folder for %s: %v", event.Title, err))
			o.stats.Errors++
			continue
		}

		for _, att := range event.Attachments {
			if att.FileID == "" {
				continue
			}

			title := att.Title
			if title == "" {
				// Fetch actual filename from Drive
				fileName, err := o.drive.GetFileName(ctx, att.FileID)
				if err != nil {
					title = fmt.Sprintf("attachment (%s...)", att.FileID[:8])
				} else {
					title = fileName
				}
			}

			result := o.drive.CreateShortcut(ctx, att.FileID, title, folder.ID, folder.Name, folder.IsNew)
			o.logCalendarAction(result, event.Title, event.Start.Format("2006-01-02"), title)
		}
	}

	return nil
}

// logActionResult logs the result of a document action.
func (o *Organizer) logActionResult(result drive.ActionResult, isMove bool) {
	if result.Skipped {
		if result.Reason == "already exists" || result.Reason == "already in folder" {
			o.stats.ItemsSkipped++
			o.logVerbose(fmt.Sprintf("   ⏭️  SKIP: %s", result.Details))
		} else if result.Reason != "" && result.Reason != "dry-run" {
			o.stats.Errors++
			o.log(fmt.Sprintf("   ⚠️  ERROR: %s\n      Reason: %s", result.Details, result.Reason))
		} else {
			// dry-run, still log what would happen
			if isMove {
				o.stats.DocumentsMoved++
				o.log(fmt.Sprintf("   📄 %s", result.Details))
			} else {
				o.stats.ShortcutsCreated++
				o.log(fmt.Sprintf("   🔗 %s", result.Details))
			}
		}
	} else {
		if isMove {
			o.stats.DocumentsMoved++
		} else {
			o.stats.ShortcutsCreated++
		}
		o.logVerbose(fmt.Sprintf("   ✓ %s", result.Details))
	}
}

// logCalendarAction logs the result of a calendar sync action.
func (o *Organizer) logCalendarAction(result drive.ActionResult, eventTitle, eventDate, attachmentName string) {
	eventContext := fmt.Sprintf("%s (%s)", eventTitle, eventDate)

	if result.Skipped {
		if result.Reason == "already exists" {
			o.stats.ItemsSkipped++
			o.logVerbose(fmt.Sprintf("   ⏭️  SKIP [%s]: %s", eventContext, result.Details))
		} else if result.Reason != "" && result.Reason != "dry-run" {
			o.stats.Errors++
			o.log(fmt.Sprintf("   ⚠️  ERROR [%s]: %s\n      Reason: %s", eventContext, result.Details, result.Reason))
		} else {
			// dry-run
			o.stats.ShortcutsCreated++
			o.log(fmt.Sprintf("   📅 EVENT: %s", eventContext))
			o.log(fmt.Sprintf("      Attachment: %s", attachmentName))
			o.log(fmt.Sprintf("      %s", result.Details))
			o.log("")
		}
	} else {
		o.stats.ShortcutsCreated++
		o.logVerbose(fmt.Sprintf("   ✓ [%s] %s", eventContext, result.Details))
	}
}

// ExtractAndCreateTasks extracts action items from documents and creates tasks.
func (o *Organizer) ExtractAndCreateTasks(ctx context.Context) error {
	// Get organized documents
	documents, err := o.drive.ListMeetingDocuments(ctx, o.config.FilenameKeywords)
	if err != nil {
		return err
	}

	o.log(fmt.Sprintf("   Scanning %d documents for action items...", len(documents)))
	o.log("")

	for _, doc := range documents {
		if err := o.processDocumentForTasks(ctx, doc); err != nil {
			o.log(fmt.Sprintf("   ⚠️  Failed to process %s: %v", doc.Name, err))
			o.stats.Errors++
		}
	}

	return nil
}

// ProcessSingleDocument extracts action items from a specific document by ID.
func (o *Organizer) ProcessSingleDocument(ctx context.Context, docID string) error {
	// Create a minimal document struct for processing
	doc := &models.Document{
		ID:   docID,
		Name: docID, // Will be updated if we can get file info
	}

	// Try to get the actual document name from Drive
	fileInfo, err := o.drive.GetFileInfo(ctx, docID)
	if err == nil && fileInfo != nil {
		doc.Name = fileInfo.Name
	}

	o.log(fmt.Sprintf("   Processing document: %s", doc.Name))
	o.log("")

	if err := o.processDocumentForTasks(ctx, doc); err != nil {
		return fmt.Errorf("failed to process document: %w", err)
	}

	return nil
}

// processDocumentForTasks processes a single document for action items.
func (o *Organizer) processDocumentForTasks(ctx context.Context, doc *models.Document) error {
	checkboxes, err := o.docs.ExtractCheckboxItems(ctx, doc.ID)
	if err != nil {
		return err
	}

	if len(checkboxes) == 0 {
		return nil
	}

	o.logVerbose(fmt.Sprintf("   📄 %s: %d checkbox items", doc.Name, len(checkboxes)))

	for _, cb := range checkboxes {
		if cb.IsProcessed {
			o.stats.ItemsSkipped++
			o.logVerbose(fmt.Sprintf("      ⏭️  Skipping (already processed): %s", truncate(cb.Text, 40)))
			continue
		}

		// Extract action item with Gemini
		response, err := o.gemini.ExtractActionItem(ctx, cb.Text)
		if err != nil {
			o.log(fmt.Sprintf("      ⚠️  Gemini extraction failed: %v", err))
			o.stats.Errors++
			continue
		}

		if response.Assignee == "" {
			o.stats.ItemsSkipped++
			o.logVerbose(fmt.Sprintf("      ⏭️  No assignee found: %s", truncate(cb.Text, 40)))
			continue
		}

		// Only create tasks for the current user (Jay Flowers)
		if !strings.Contains(strings.ToLower(response.Assignee), "jay") {
			o.stats.ItemsSkipped++
			o.logVerbose(fmt.Sprintf("      ⏭️  Assigned to others: %s -> %s", truncate(cb.Text, 30), response.Assignee))
			continue
		}

		dueDate, _ := gemini.ParseDate(response.Date)
		dateStr := response.Date
		if dateStr == "" {
			dateStr = "(no date)"
		}

		actionItem := &models.ActionItem{
			DocumentID:   doc.ID,
			DocumentName: doc.Name,
			Text:         cb.Text,
			Assignee:     response.Assignee,
			DueDate:      dueDate,
			LineIndex:    int(cb.EndIndex - 1),
		}

		if o.drive.IsDryRun() {
			o.stats.TasksCreated++
			o.log(fmt.Sprintf("   ✓ Created task: %s -> %s", truncate(cb.Text, 30), response.Assignee))
			continue
		}

		// Create the task
		taskID, err := o.tasks.CreateTask(ctx, actionItem)
		if err != nil {
			o.log(fmt.Sprintf("      ⚠️  Failed to create task: %v", err))
			o.stats.Errors++
			continue
		}
		actionItem.TaskID = taskID
		o.stats.TasksCreated++

		// Note: We no longer annotate the document - just create the task
		o.log(fmt.Sprintf("   ✓ Created task: %s -> %s", truncate(cb.Text, 30), response.Assignee))
	}

	return nil
}

func (o *Organizer) log(msg string) {
	fmt.Println(msg)
}

func (o *Organizer) logVerbose(msg string) {
	if o.verbose {
		fmt.Println(msg)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
