# Feature Specification: Meeting Organizer & Action Item Tracker CLI

**Feature Branch**: `001-gcal-organizer-cli`  
**Created**: 2026-02-01  
**Status**: Draft  
**Input**: Rewrite of Google Apps Script to Go CLI with GCP Gemini API key support

## Overview

A command-line tool that automates the lifecycle of meeting notes by:
1. Organizing meeting documents into a structured Drive folder hierarchy
2. Syncing calendar event attachments to corresponding meeting folders
3. Using Gemini AI to extract action items from checkboxes in Google Docs
4. Creating Google Tasks from extracted action items

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Organize Meeting Documents (Priority: P1)

As a user, I want to run a command that scans my Google Drive for meeting notes and automatically organizes them into topic-based subfolders within my Master Folder.

**Why this priority**: This is the core organizational function. Without it, the other features have no context.

**Independent Test**: Can be fully tested by creating a test document with a meeting name pattern and verifying it moves to the correct subfolder.

**Acceptance Scenarios**:

1. **Given** a document named "Team Standup - 2026-01-30" exists in Drive, **When** I run `gcal-organizer organize`, **Then** a subfolder "Team Standup" is created (or located) in the Master Folder and the document is moved there.

2. **Given** a document I don't own named "External Meeting - 2026-01-30", **When** I run `gcal-organizer organize`, **Then** a shortcut to the document is created in the "External Meeting" subfolder instead of moving it.

3. **Given** a document that doesn't match the naming pattern (e.g., "Random Notes"), **When** I run `gcal-organizer organize`, **Then** the document is skipped and logged appropriately.

---

### User Story 2 - Sync Calendar Attachments (Priority: P2)

As a user, I want the tool to scan my recent calendar events and sync any attached documents or Drive links to the corresponding meeting folders.

**Why this priority**: Extends organization capability to calendar-linked resources.

**Independent Test**: Can be tested by creating a calendar event with an attachment and running the sync command.

**Acceptance Scenarios**:

1. **Given** a calendar event "Weekly Planning" with a Drive attachment, **When** I run `gcal-organizer sync-calendar --days 8`, **Then** a shortcut to the attachment appears in the "Weekly Planning" subfolder.

2. **Given** a calendar event with a Drive URL in the description (not as attachment), **When** I run `gcal-organizer sync-calendar`, **Then** the URL is detected and a shortcut is created.

3. **Given** a calendar event with no attachments or Drive links, **When** I run `gcal-organizer sync-calendar`, **Then** the event is skipped gracefully.

---

### User Story 3 - Extract Action Items with AI (Priority: P2)

As a user, I want the tool to scan Google Docs for checkbox list items and use Gemini AI to extract assignees and due dates.

**Why this priority**: Core AI-powered automation feature that differentiates this tool.

**Independent Test**: Can be tested with a sample document containing checkboxes with task descriptions.

**Acceptance Scenarios**:

1. **Given** a Google Doc with unchecked checkbox item "[ ] John to review budget by Friday", **When** I run `gcal-organizer extract-tasks --doc <docId>`, **Then** Gemini returns JSON with `{"assignee": "John", "date": "2026-02-07"}`.

2. **Given** a checkbox item already marked with 🆔 emoji, **When** I run `gcal-organizer extract-tasks`, **Then** the item is skipped (already processed).

3. **Given** a checkbox item with ambiguous text "[ ] Do the thing", **When** Gemini cannot determine assignee, **Then** the item is logged but no task is created.

---

### User Story 4 - Create Google Tasks (Priority: P3)

As a user, I want extracted action items to be automatically created as Google Tasks with proper titles and due dates.

**Why this priority**: Completes the automation loop but depends on successful extraction.

**Independent Test**: Can be tested by mocking Gemini response and verifying Task creation.

**Acceptance Scenarios**:

1. **Given** Gemini returns `{"assignee": "John", "date": "2026-02-07"}` for a checkbox item, **When** processing completes, **Then** a Google Task is created with title "[Doc Name] Review budget" and due date 2026-02-07.

2. **Given** a task is successfully created, **When** the doc is updated, **Then** the checkbox line is annotated with `👤 John 📅 2026-02-07 🆔`.

---

### User Story 5 - Run Full Workflow (Priority: P1)

As a user, I want a single command to run the complete workflow: organize, sync, extract, and create tasks.

**Why this priority**: Primary user interaction mode for daily use.

**Independent Test**: Can be tested with a sample setup of docs and calendar events.

**Acceptance Scenarios**:

1. **Given** configured environment variables and credentials, **When** I run `gcal-organizer run`, **Then** all sub-workflows execute in sequence with progress output.

2. **Given** `--dry-run` flag is provided, **When** I run `gcal-organizer run --dry-run`, **Then** all actions are logged but no changes are made to Drive, Docs, or Tasks.

---

### Edge Cases

- What happens when the GCP API key is invalid or missing?
  → Clear error message with setup instructions
- What happens when OAuth token is expired?
  → Prompt for re-authentication or provide refresh instructions
- What happens when rate limits are exceeded?
  → Implement exponential backoff with clear status messages
- What happens when a document is trashed during processing?
  → Log the error and continue with remaining documents
- What happens when Gemini returns malformed JSON?
  → Parse gracefully, log the error, skip the item

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST authenticate with Google Workspace APIs using OAuth2 (Drive, Docs, Calendar, Tasks)
- **FR-002**: System MUST authenticate with Gemini API using GCP API key
- **FR-003**: System MUST parse document names using configurable regex pattern (default: `(.+) - (\d{4}-\d{2}-\d{2})`)
- **FR-004**: System MUST create subfolders based on extracted meeting names
- **FR-005**: System MUST move owned documents and create shortcuts for non-owned documents
- **FR-006**: System MUST scan calendar events within configurable lookback period (default: 8 days)
- **FR-007**: System MUST detect Drive attachments via Calendar API and URLs in event descriptions
- **FR-008**: System MUST identify checkbox list items in Google Docs
- **FR-009**: System MUST skip checkbox items already containing 🆔 emoji
- **FR-010**: System MUST send checkbox text to Gemini and parse JSON response for assignee and date
- **FR-011**: System MUST create Google Tasks with formatted titles and due dates
- **FR-012**: System MUST annotate processed checkbox items with `👤 Name 📅 Date 🆔`
- **FR-013**: System MUST support `--dry-run` mode with detailed output showing:
  - **FR-013a**: For document organization: file name, source location, target folder, and action (move vs shortcut)
  - **FR-013b**: For calendar sync: event title, attachment name, target folder for shortcut
  - **FR-013c**: For task extraction: checkbox text, extracted assignee, extracted date, target document
  - **FR-013d**: Summary counts at the end (documents processed, shortcuts to create, tasks to create)
- **FR-014**: System MUST support `--verbose` mode for detailed logging

### Configuration Requirements

- **CR-001**: `GCAL_MASTER_FOLDER_NAME` - Name of the master folder (string)
- **CR-002**: `GCAL_DAYS_TO_LOOK_BACK` - Calendar lookback period (integer, default: 8)
- **CR-003**: `GEMINI_API_KEY` - GCP API key for Gemini (string, required)
- **CR-004**: `GCAL_FILENAME_KEYWORDS` - Keywords to filter documents (comma-separated list)
- **CR-005**: `GCAL_FILENAME_PATTERN` - Regex for parsing document names (string)

### Key Entities

- **Meeting Folder**: A subfolder within the Master Folder, named after a meeting topic
- **Meeting Document**: A Google Doc matching the naming pattern
- **Calendar Event**: A Google Calendar event with potential attachments
- **Action Item**: A checkbox item in a document with extractable assignee/date
- **Task**: A Google Task created from an action item

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: CLI responds to all commands within 2 seconds (excluding API calls)
- **SC-002**: Successfully organizes 100+ documents without errors in a single run
- **SC-003**: Gemini extraction accuracy > 85% for well-formatted checkbox items
- **SC-004**: All operations are idempotent (safe to run multiple times)
- **SC-005**: Error messages are actionable (include next steps or troubleshooting hints)
- **SC-006**: `--dry-run` mode produces identical log output without side effects

## Clarifications

*To be filled in during /speckit.clarify phase*

## Review & Acceptance Checklist

- [ ] All user stories have acceptance scenarios
- [ ] All functional requirements are testable
- [ ] Edge cases are documented
- [ ] Configuration requirements are complete
- [ ] Success criteria are measurable
- [ ] Dependencies (OAuth, API key) are documented
