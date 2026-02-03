# GCal Organizer

A Go CLI tool that automates meeting note organization, calendar attachment syncing, and AI-powered action item extraction using Google Workspace APIs and Gemini AI.

## 🚀 Quick Start

This project uses [Spec-Kit](https://github.com/github/spec-kit) for Spec-Driven Development. The `/speckit.*` commands are available in your AI agent (Gemini CLI).

### Prerequisites

- Go 1.21+
- GCP API Key for Gemini (set as `GEMINI_API_KEY`)
- Google Cloud OAuth2 credentials for Workspace APIs

### Spec-Driven Development Workflow

1. **Review the Constitution**: `.specify/memory/constitution.md`
2. **Review the Specification**: `.specify/specs/001-gcal-organizer-cli/spec.md`
3. **Create a Plan**: Run `/speckit.plan` in your AI agent
4. **Generate Tasks**: Run `/speckit.tasks` in your AI agent  
5. **Implement**: Run `/speckit.implement` in your AI agent

## 📁 Project Structure

```
gcal-organizer/
├── cmd/gcal-organizer/
│   └── main.go              # CLI entry point (Cobra)
├── internal/
│   ├── auth/
│   │   ├── oauth.go         # OAuth2 for Workspace APIs
│   │   └── gemini.go        # Gemini API key auth
│   ├── calendar/
│   │   └── service.go       # Calendar event operations
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── docs/
│   │   └── service.go       # Docs checkbox extraction
│   ├── drive/
│   │   └── service.go       # Drive folder/file operations  
│   ├── gemini/
│   │   └── client.go        # Gemini AI extraction
│   ├── organizer/
│   │   └── organizer.go     # Workflow orchestration
│   └── tasks/
│       └── service.go       # Google Tasks creation
├── pkg/models/
│   └── models.go            # Shared data structures
├── .specify/                # Spec-kit artifacts
├── .env.example             # Environment config template
├── .go-version              # Go version (1.22.4)
├── Makefile                 # Build/test commands
├── go.mod / go.sum          # Go dependencies
└── README.md
```

## 🔑 Configuration

Set the following environment variables:

```bash
# Required: GCP API Key for Gemini
export GEMINI_API_KEY="your-gcp-api-key"

# Optional: Configuration
export GCAL_MASTER_FOLDER_NAME="Meeting Notes"
export GCAL_DAYS_TO_LOOK_BACK="8"
export GCAL_FILENAME_KEYWORDS="Notes,Meeting"
```

## 📋 Features (To Be Implemented)

1. **Organize Documents** - Auto-organize meeting notes into topic folders
2. **Sync Calendar** - Link calendar attachments to meeting folders  
3. **Extract Tasks** - Use Gemini AI to extract action items from docs
4. **Create Tasks** - Auto-create Google Tasks with due dates

## 🤖 Using Spec-Kit Commands

In your AI agent (Gemini CLI), use these commands:

| Command | Description |
|---------|-------------|
| `/speckit.constitution` | View or update project principles |
| `/speckit.specify` | Create or update feature specifications |
| `/speckit.clarify` | Ask questions to clarify ambiguous requirements |
| `/speckit.plan` | Create technical implementation plan |
| `/speckit.tasks` | Generate actionable task breakdown |
| `/speckit.implement` | Execute implementation based on tasks |
| `/speckit.analyze` | Check consistency across artifacts |
| `/speckit.checklist` | Generate quality validation checklists |

## 📄 License

See [LICENSE](LICENSE) for details.
