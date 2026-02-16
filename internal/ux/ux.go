// Package ux provides user-facing error types with actionable fix suggestions.
package ux

import "fmt"

// ActionError is an error that includes a fix suggestion for the user.
type ActionError struct {
	Message string
	Fix     string
}

func (e *ActionError) Error() string {
	if e.Fix == "" {
		return fmt.Sprintf("❌ Error: %s", e.Message)
	}
	return fmt.Sprintf("❌ Error: %s\n   Fix:   %s", e.Message, e.Fix)
}

// NewError creates an ActionError with a message and fix.
func NewError(message, fix string) *ActionError {
	return &ActionError{Message: message, Fix: fix}
}

// --- Common error constructors ---

// MissingCredentials returns an error for missing OAuth credentials file.
func MissingCredentials(path string) *ActionError {
	return &ActionError{
		Message: fmt.Sprintf("credentials.json not found at %s", path),
		Fix:     "Run 'gcal-organizer init' or download from https://console.cloud.google.com/apis/credentials",
	}
}

// MissingAPIKey returns an error for missing Gemini API key.
func MissingAPIKey() *ActionError {
	return &ActionError{
		Message: "GEMINI_API_KEY is not set",
		Fix:     "Set GEMINI_API_KEY in ~/.gcal-organizer/.env or run 'gcal-organizer init'",
	}
}

// TokenExpired returns an error for expired OAuth token.
func TokenExpired() *ActionError {
	return &ActionError{
		Message: "OAuth token is expired or invalid",
		Fix:     "Run 'gcal-organizer auth login' to re-authenticate",
	}
}

// MissingToken returns an error for missing OAuth token.
func MissingToken() *ActionError {
	return &ActionError{
		Message: "not authenticated — no OAuth token found",
		Fix:     "Run 'gcal-organizer auth login' to authenticate with Google",
	}
}

// MissingNodeJS returns an error for missing Node.js.
func MissingNodeJS() *ActionError {
	return &ActionError{
		Message: "Node.js not found (required for task assignment)",
		Fix:     "Install Node.js 18+ from https://nodejs.org — Steps 1-2 still work without it",
	}
}

// MissingChromeProfile returns an error for missing Chrome data directory.
func MissingChromeProfile(path string) *ActionError {
	return &ActionError{
		Message: fmt.Sprintf("Chrome data directory not found at %s", path),
		Fix:     "Run 'gcal-organizer setup-browser' to create it",
	}
}

// MissingConfigDir returns an error for missing config directory.
func MissingConfigDir() *ActionError {
	return &ActionError{
		Message: "config directory ~/.gcal-organizer/ does not exist",
		Fix:     "Run 'gcal-organizer init' to set up your configuration",
	}
}

// MissingEnvFile returns an error for missing .env file.
func MissingEnvFile() *ActionError {
	return &ActionError{
		Message: "environment file ~/.gcal-organizer/.env not found",
		Fix:     "Run 'gcal-organizer init' to generate your configuration",
	}
}

// ServiceNotInstalled returns an error for missing service.
func ServiceNotInstalled() *ActionError {
	return &ActionError{
		Message: "hourly service is not installed",
		Fix:     "Run 'gcal-organizer install' to set up the hourly service",
	}
}
