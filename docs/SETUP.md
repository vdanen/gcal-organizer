# Setup Guide

This guide walks you through setting up `gcal-organizer` with Google Cloud credentials.

## Prerequisites

- Go 1.21 or later
- A Google account with access to:
  - Google Drive
  - Google Calendar
  - Google Docs
  - Google Tasks
- A Google Cloud project with billing enabled (for Gemini API)

---

## Step 1: Create a Google Cloud Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Click "Select a project" → "New Project"
3. Name it (e.g., `gcal-organizer`) and click "Create"
4. Select the new project

---

## Step 2: Enable Required APIs

Navigate to **APIs & Services** → **Library** and enable:

| API | Purpose |
|-----|---------|
| Google Drive API | Organize documents, create folders/shortcuts |
| Google Calendar API | Read calendar events and attachments |
| Google Docs API | Read document content for action items |
| Google Tasks API | Create tasks from extracted action items |
| Generative Language API | Gemini AI for parsing action items |

**Quick links:**
- [Enable Drive API](https://console.cloud.google.com/apis/library/drive.googleapis.com)
- [Enable Calendar API](https://console.cloud.google.com/apis/library/calendar-json.googleapis.com)
- [Enable Docs API](https://console.cloud.google.com/apis/library/docs.googleapis.com)
- [Enable Tasks API](https://console.cloud.google.com/apis/library/tasks.googleapis.com)
- [Enable Generative Language API](https://console.cloud.google.com/apis/library/generativelanguage.googleapis.com)

---

## Step 3: Create OAuth 2.0 Credentials

1. Go to **APIs & Services** → **Credentials**
2. Click **Create Credentials** → **OAuth client ID**
3. If prompted, configure the OAuth consent screen:
   - User Type: **External** (or Internal if using Workspace)
   - App name: `gcal-organizer`
   - User support email: Your email
   - Developer contact: Your email
   - Scopes: Skip for now (we'll use the default scopes)
   - Test users: Add your email
4. Back in Credentials, create OAuth client ID:
   - Application type: **Desktop app**
   - Name: `gcal-organizer-cli`
5. Click **Create** and download the JSON file
6. Save it as `~/.gcal-organizer/credentials.json`:

```bash
mkdir -p ~/.gcal-organizer
mv ~/Downloads/client_secret_*.json ~/.gcal-organizer/credentials.json
```

---

## Step 4: Get a Gemini API Key

1. Go to [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Click **Create API Key**
3. Select your project or create a new one
4. Copy the API key

---

## Step 5: Configure the Application

Create a configuration file at `~/.gcal-organizer/config.yaml`:

```yaml
# Required: Your Gemini API key
gemini_api_key: "YOUR_GEMINI_API_KEY_HERE"

# Master folder name in Google Drive (will be created if it doesn't exist)
master_folder_name: "Meeting Notes"

# Keywords to identify meeting documents (regex patterns)
filename_keywords:
  - "Notes"
  - "Meeting"
  - "Standup"
  - "1-1"
  - "Weekly"

# Days to look back for calendar events
calendar_lookback_days: 8

# Task list name in Google Tasks
task_list_name: "Meeting Action Items"

# Optional: Enable verbose output
verbose: false
```

Alternatively, use environment variables (see `.env.example`):

```bash
export GEMINI_API_KEY="your-key-here"
export MASTER_FOLDER_NAME="Meeting Notes"
```

---

## Step 6: Build and Authenticate

```bash
# Clone and build
git clone https://github.com/jflowers/gcal-organizer.git
cd gcal-organizer
go build -o gcal-organizer ./cmd/gcal-organizer

# First run - this will open a browser for OAuth
./gcal-organizer auth login
```

The OAuth flow will:
1. Open your browser to Google's consent page
2. Ask you to authorize the app
3. Store the token at `~/.gcal-organizer/token.json`

---

## Step 7: Verify Setup

```bash
# Check configuration
./gcal-organizer config show

# Test with dry-run (no changes made)
./gcal-organizer run --dry-run --verbose
```

---

## Scopes Requested

The app requests these OAuth scopes:

| Scope | Purpose |
|-------|---------|
| `drive.file` | Create folders and shortcuts (only accesses files created by the app) |
| `drive.readonly` | Read file metadata for organizing |
| `calendar.readonly` | Read calendar events and attachments |
| `documents.readonly` | Read document content for action items |
| `tasks` | Create and manage tasks |

---

## Troubleshooting

### "Access blocked: This app's request is invalid"

Your OAuth consent screen may not be configured correctly:
1. Go to **APIs & Services** → **OAuth consent screen**
2. Make sure your email is added as a **Test user**
3. Or publish the app (requires verification for sensitive scopes)

### "File not found" errors for calendar attachments

This usually means:
- The file was deleted
- You don't have access to the file
- The file is in someone else's Drive and not shared with you

These are expected for external meeting recordings/transcripts.

### "Invalid credentials" error

Your `credentials.json` may be corrupted or wrong type:
1. Delete `~/.gcal-organizer/credentials.json`
2. Re-download from Google Cloud Console
3. Make sure it's a **Desktop app** credential, not Web or Service Account

### "Token expired" or authentication issues

Delete the token and re-authenticate:
```bash
rm ~/.gcal-organizer/token.json
./gcal-organizer auth login
```

---

## Security Notes

- **Never commit credentials**: `credentials.json` and `token.json` are in `.gitignore`
- **Token storage**: Tokens are stored locally at `~/.gcal-organizer/token.json`
- **Minimal scopes**: The app requests only the scopes it needs
- **Offline access**: The token includes refresh capability for long-running use

---

## Next Steps

Once setup is complete:

```bash
# Run the full workflow
./gcal-organizer run --verbose

# Or run individual steps
./gcal-organizer organize --dry-run
./gcal-organizer sync-calendar --days 14
./gcal-organizer extract-tasks --doc <DOC_ID>
```

See the [README](../README.md) for full usage documentation.
