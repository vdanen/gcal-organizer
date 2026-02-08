# Homebrew Tap: gcal-organizer

Homebrew formulae for [gcal-organizer](https://github.com/jflowers/gcal-organizer) — automate meeting note organization, calendar syncing, and task assignment.

## Installation

```bash
brew tap jflowers/gcal-organizer
brew install gcal-organizer
```

## Quick Start

```bash
gcal-organizer init           # Set up configuration
gcal-organizer auth login     # Authenticate with Google
gcal-organizer setup-browser  # Configure browser automation
gcal-organizer doctor         # Verify everything is set up
gcal-organizer run --dry-run  # Test without making changes
gcal-organizer install        # Install hourly service
```

## Updating

```bash
brew update
brew upgrade gcal-organizer
```

## Supported Platforms

- **macOS** (Apple Silicon & Intel)
- **Linux** (x86_64 & arm64) via [Linuxbrew](https://docs.brew.sh/Homebrew-on-Linux)
