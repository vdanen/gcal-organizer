// Package organizer provides the main orchestration logic.
package organizer

import (
	"context"
	"fmt"
	"strings"

	"github.com/jflowers/gcal-organizer/internal/calendar"
	"github.com/jflowers/gcal-organizer/internal/config"
	"github.com/jflowers/gcal-organizer/internal/drive"
)

// Stats tracks operation counts for summary reporting.
type Stats struct {
	DocumentsFound     int
	DocumentsMoved     int
	ShortcutsCreated   int
	ShortcutsTrashed   int
	EventsProcessed    int
	EventsWithAttach   int
	FoldersShared      int
	AttachmentsShared  int
	TasksAssigned      int
	TasksFailed        int
	Errors             int
}

// Organizer orchestrates all the services.
type Organizer struct {
	config   *config.Config
	drive    *drive.Service
	calendar *calendar.Service

	verbose     bool
	stats       Stats
	notesDocIDs map[string]bool // Google Doc IDs with "Notes" attachments
}

// New creates a new Organizer with all services initialized.
func New(cfg *config.Config, driveSvc *drive.Service, calSvc *calendar.Service) *Organizer {
	return &Organizer{
		config:      cfg,
		drive:       driveSvc,
		calendar:    calSvc,
		verbose:     cfg.Verbose,
		notesDocIDs: make(map[string]bool),
	}
}

// GetNotesDocIDs returns the list of Google Doc IDs with "Notes" attachments.
func (o *Organizer) GetNotesDocIDs() []string {
	var ids []string
	for id := range o.notesDocIDs {
		ids = append(ids, id)
	}
	return ids
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

	// Note: Step 3 (Assign Tasks) is handled by the caller if needed,
	// since it requires browser automation that lives outside the organizer.

	return nil
}

// PrintSummary outputs the final statistics.
func (o *Organizer) PrintSummary() {
	o.printSummary()
}

// AddTaskStats updates the task assignment statistics.
func (o *Organizer) AddTaskStats(assigned, failed int) {
	o.stats.TasksAssigned += assigned
	o.stats.TasksFailed += failed
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
	if o.stats.FoldersShared > 0 {
		o.log(fmt.Sprintf("   👥 Folders shared:          %d", o.stats.FoldersShared))
	}
	if o.stats.AttachmentsShared > 0 {
		o.log(fmt.Sprintf("   📎 Attachments shared:      %d", o.stats.AttachmentsShared))
	}
	o.log(fmt.Sprintf("   ✅ Tasks assigned:          %d", o.stats.TasksAssigned))
	if o.stats.TasksFailed > 0 {
		o.log(fmt.Sprintf("   ❌ Tasks failed:            %d", o.stats.TasksFailed))
	}
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

			// Track Google Docs with "Notes" in the title for task assignment
			if att.MimeType == "application/vnd.google-apps.document" &&
				strings.Contains(strings.ToLower(title), "notes") {
				o.notesDocIDs[att.FileID] = true
			}
		}

		// Share folder with attendees
		for _, attendee := range event.Attendees {
			// Skip self, empty emails, and calendar resources
			if attendee.IsSelf || attendee.Email == "" || isCalendarResource(attendee.Email) {
				continue
			}

			result := o.drive.ShareFolder(ctx, folder.ID, folder.Name, attendee.Email)
			if !result.Skipped || result.Reason == "dry-run" {
				o.stats.FoldersShared++
				o.log(fmt.Sprintf("   👥 %s", result.Details))
			} else if result.Reason != "already shared" {
				// Log errors but not "already shared" skips
				o.log(fmt.Sprintf("   ⚠️  %s", result.Details))
				o.stats.Errors++
			}
		}

		// Share attachments with attendees (edit access)
		for _, att := range event.Attachments {
			if att.FileID == "" {
				continue
			}

			// Only share if we have edit access to the attachment
			if !o.drive.CanEditFile(ctx, att.FileID) {
				o.logVerbose(fmt.Sprintf("   ⏭️  SKIP: No edit access to '%s'", att.Title))
				continue
			}

			for _, attendee := range event.Attendees {
				if attendee.IsSelf || attendee.Email == "" || isCalendarResource(attendee.Email) {
					continue
				}

				result := o.drive.ShareFile(ctx, att.FileID, att.Title, attendee.Email, "writer")
				if !result.Skipped || result.Reason == "dry-run" {
					o.stats.AttachmentsShared++
					o.log(fmt.Sprintf("   📎 %s", result.Details))
				} else if result.Reason != "already shared" {
					o.log(fmt.Sprintf("   ⚠️  %s", result.Details))
					o.stats.Errors++
				}
			}
		}
	}

	return nil
}

// isCalendarResource returns true for Google Calendar resource/group addresses
// that cannot be shared with (e.g. conference rooms, group calendars).
func isCalendarResource(email string) bool {
	return strings.HasSuffix(email, "@resource.calendar.google.com") ||
		strings.HasSuffix(email, "@group.calendar.google.com")
}

// logActionResult logs the result of a document action.
func (o *Organizer) logActionResult(result drive.ActionResult, isMove bool) {
	if result.Skipped {
		if result.Reason == "already exists" || result.Reason == "already in folder" {
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

func (o *Organizer) log(msg string) {
	fmt.Println(msg)
}

func (o *Organizer) logVerbose(msg string) {
	if o.verbose {
		fmt.Println(msg)
	}
}
