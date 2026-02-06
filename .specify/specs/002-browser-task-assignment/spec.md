# Feature Specification: Browser Automation for Task Assignment

**Feature Branch**: `002-browser-task-assignment`  
**Created**: 2026-02-06  
**Status**: Draft  
**Input**: User wants to automate native Google Docs checkbox task assignment using browser automation

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Assign Tasks to Anyone (Priority: P1)

As a user running `gcal-organizer assign-tasks --doc <id>`, I want the system to automatically assign checkbox items in the "Suggested next steps" section to the appropriate people using Google Docs' native UI.

**Why this priority**: Core functionality - the main reason for this feature

**Independent Test**: Can be tested by running against a specific document and visually confirming assignments appear with user avatars in the doc

**Acceptance Scenarios**:

1. **Given** a Google Doc with checkbox items in "Suggested next steps" mentioning "Jay Flowers", **When** I run `assign-tasks --doc <id>`, **Then** the checkbox is assigned to Jay Flowers via the native UI and shows their avatar
2. **Given** a checkbox mentioning multiple people, **When** I run `assign-tasks`, **Then** the system uses Gemini to determine the primary assignee
3. **Given** a checkbox already assigned, **When** I run `assign-tasks`, **Then** the system skips that checkbox

---

### User Story 2 - Preview Mode (Priority: P2)

As a user, I want to preview what assignments would be made before the browser automation runs.

**Why this priority**: Safety feature - allows verification before making changes

**Independent Test**: Run with `--dry-run` flag and verify output shows planned assignments without opening browser

**Acceptance Scenarios**:

1. **Given** a document with assignable tasks, **When** I run `assign-tasks --doc <id> --dry-run`, **Then** I see a list of planned assignments without any browser interaction
2. **Given** dry-run output, **When** user reviews, **Then** they see assignee emails and task descriptions

---

### User Story 3 - Progress Feedback (Priority: P3)

As a user, I want to see progress as tasks are being assigned.

**Why this priority**: UX improvement - helps user understand what's happening

**Independent Test**: Run command and observe console output during execution

**Acceptance Scenarios**:

1. **Given** browser automation is running, **When** each task is assigned, **Then** output shows "✓ Assigned: [task] → [email]"
2. **Given** an assignment fails, **When** error occurs, **Then** output shows "✗ Failed: [task] - [reason]"

---

### Edge Cases

- What happens when the assignee email is not recognized by Google Docs?
- How does system handle checkboxes that are already assigned?
- What if the document is open in another browser/user?
- How to handle rate limiting or slow network?
- What if "Suggested next steps" section doesn't exist?
- What happens when Gemini returns malformed JSON?
  → Parse gracefully, log the error, skip the item

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST open a browser and navigate to the target Google Doc
- **FR-002**: System MUST use existing Chrome profile for Google authentication
- **FR-003**: System MUST locate the "Suggested next steps" section in the document
- **FR-004**: System MUST identify unchecked checkbox items in that section
- **FR-005**: System MUST use Gemini to extract assignee names from each checkbox text
- **FR-006**: System MUST resolve assignee names to email addresses via the document's attendee list (from the "Invited" header)
- **FR-007**: System MUST click each checkbox to open the assignment popup
- **FR-008**: System MUST enter the assignee email in the "Assignee" field
- **FR-009**: System MUST click "Assign as a task" button
- **FR-010**: System MUST skip checkboxes that already have an assignee avatar
- **FR-011**: System MUST provide `--dry-run` mode to preview without browser
- **FR-012**: System MUST provide verbose output showing each step

### Configuration Requirements

- **CR-001**: `GEMINI_API_KEY` - GCP API key for Gemini (string, required)
- **CR-002**: `CHROME_PROFILE_PATH` - Path to Chrome profile directory for authentication (default: `/Users/jflowers/Library/Application Support/Google/Chrome/Profile 1`)

### Key Entities

- **TaskAssignment**: Checkbox text, assignee name (from Gemini), assignee email, status
- **DocumentContext**: Doc ID, title, attendee list with emails

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `assign-tasks --doc <id>` successfully assigns all identifiable tasks
- **SC-002**: Assignee avatars appear next to checkboxes in the document
- **SC-003**: Assigned users receive task notifications (via Google's native system)
- **SC-004**: Dry-run mode shows accurate preview of planned assignments
- **SC-005**: Command completes within 30 seconds for a document with 10 tasks
- **SC-006**: Gemini extraction accuracy > 85% for well-formatted checkbox items
