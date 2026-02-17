# GCal Organizer Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-17

## Active Technologies

- **Language**: Go 1.21+
- **CLI Framework**: github.com/spf13/cobra
- **Google APIs**: Drive v3, Docs v1, Calendar v3, Tasks v1
- **AI**: Gemini API via google.golang.org/genai
- **Browser Automation**: Playwright (TypeScript) via npx tsx
- **Authentication**: OAuth2 (Workspace), GCP API Key (Gemini)

## Project Structure

```text
gcal-organizer/
├── cmd/gcal-organizer/          # CLI entry point
├── internal/
│   ├── auth/                    # OAuth2 and API key handling
│   ├── config/                  # Configuration management
│   ├── drive/                   # Google Drive operations
│   ├── docs/                    # Google Docs parsing
│   ├── calendar/                # Calendar operations
│   ├── tasks/                   # Tasks API operations
│   ├── gemini/                  # Gemini AI client
│   └── organizer/               # Main orchestration
├── pkg/models/                  # Shared data models
├── scripts/                     # Browser automation (TypeScript)
├── .specify/                    # Spec-driven development artifacts
└── .opencode/                   # OpenCode agent commands
```

## Commands

```bash
# Build
go build ./...

# Test
go test ./...

# Lint
go vet ./...
gofmt -l .

# Run CI checks locally
make ci

# Install git hooks
make install-hooks
```

## Code Style

### Go Conventions
- Standard project layout (cmd/, internal/, pkg/)
- Error handling via explicit return values, not panic
- Use `context.Context` for cancellation and timeouts
- Table-driven tests preferred
- Wrap errors with context using `fmt.Errorf` with `%w`

### Documentation
- README.md, SETUP.md, man pages must be kept current
- New features require documentation before completion

## Recent Changes

### 001-gcal-organizer-cli
Core CLI implementation with Google Workspace integration and Gemini AI for action item extraction.

### 002-browser-task-assignment
Browser automation via Playwright for task assignment through Google Docs native UI.

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
