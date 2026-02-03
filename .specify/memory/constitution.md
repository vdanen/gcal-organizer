# GCal Organizer Constitution

## Core Principles

### I. CLI-First Architecture
All functionality is exposed through a command-line interface.
- Commands follow standard Unix conventions (stdin/stdout/stderr)
- Support both human-readable and JSON output formats
- Exit codes follow POSIX conventions (0 = success, non-zero = error)
- Configuration via environment variables and config files

### II. API-Key Authentication Mode
Uses GCP API key for Gemini access (not OAuth for Gemini).
- Google Workspace APIs (Drive, Docs, Calendar, Tasks) use OAuth2
- Gemini API uses GCP API key via Vertex AI compatible endpoint
- Credentials stored securely via environment variables
- Never log or expose credentials in output

### III. Test-Driven Development
Testing is mandatory before implementation.
- Unit tests for all business logic
- Integration tests for Google API interactions (use mocks)
- Table-driven tests preferred for Go
- `go test` must pass before any commit

### IV. Idiomatic Go
Follow Go best practices and conventions.
- Standard project layout (cmd/, internal/, pkg/)
- Error handling via explicit return values, not panic
- Use `context.Context` for cancellation and timeouts
- Minimize external dependencies; prefer standard library

### V. Graceful Error Handling
User-friendly error messages with actionable guidance.
- Wrap errors with context using `fmt.Errorf` with `%w`
- Log sufficient detail for debugging
- Display human-readable messages to users
- Support verbose/debug mode for troubleshooting

### VI. Observability
Built-in logging and debugging capabilities.
- Structured logging with configurable levels
- Dry-run mode for testing without side effects
- Progress indicators for long-running operations
- Clear output of what actions were taken

## Technical Constraints

### API Limitations
- GCP Gemini API key (project: gcp-jflowers-pro-gemini)
- Use `google.golang.org/genai` SDK
- Model: gemini-1.5-flash (or configurable)
- Handle rate limiting and quotas gracefully

### Google Workspace Integration
Required scopes:
- `https://www.googleapis.com/auth/documents` (Read/Write)
- `https://www.googleapis.com/auth/drive` (File operations)
- `https://www.googleapis.com/auth/calendar.readonly` (Event scanning)
- `https://www.googleapis.com/auth/tasks` (Task creation)

### Technology Stack
- Language: Go 1.21+
- Build: Standard Go toolchain
- Dependencies: Go modules
- No CGO required (pure Go for portability)

## Development Workflow

### Code Review Requirements
- All changes via feature branches
- Clear commit messages following conventional commits
- Tests must pass before merge
- Self-documented code preferred over comments

### Quality Gates
- `go build ./...` - must compile
- `go test ./...` - must pass
- `go vet ./...` - must have no warnings
- `gofmt` - code must be formatted

## Governance

This constitution guides all development decisions for the GCal Organizer CLI.
- Amendments require documentation and justification
- Trade-offs against principles must be explicitly noted
- Simplicity preferred over feature creep (YAGNI)

**Version**: 1.0.0 | **Ratified**: 2026-02-01 | **Last Amended**: 2026-02-01
