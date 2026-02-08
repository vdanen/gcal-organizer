package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// doctorCmd checks system health and reports issues with fixes.
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and report issues with fixes",
	Long:  `Diagnose common issues with gcal-organizer setup and report actionable fixes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		configDir := filepath.Join(home, ".gcal-organizer")
		passed := 0
		warned := 0
		failed := 0

		fmt.Println("🩺 gcal-organizer doctor")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println()

		// 1. Config directory
		if info, err := os.Stat(configDir); err == nil && info.IsDir() {
			fmt.Println("  ✅ Config directory exists")
			passed++
		} else {
			fmt.Println("  ❌ Config directory ~/.gcal-organizer/ not found")
			fmt.Println("     Fix: Run 'gcal-organizer init'")
			failed++
		}

		// 2. .env file
		envFile := filepath.Join(configDir, ".env")
		if _, err := os.Stat(envFile); err == nil {
			fmt.Println("  ✅ Environment file (.env) exists")
			passed++
		} else {
			fmt.Println("  ❌ Environment file ~/.gcal-organizer/.env not found")
			fmt.Println("     Fix: Run 'gcal-organizer init'")
			failed++
		}

		// 3. credentials.json
		credFile := filepath.Join(configDir, "credentials.json")
		if _, err := os.Stat(credFile); err == nil {
			fmt.Println("  ✅ Google credentials (credentials.json) found")
			passed++
		} else {
			fmt.Println("  ❌ Google credentials not found at", credFile)
			fmt.Println("     Fix: Download from https://console.cloud.google.com/apis/credentials")
			fmt.Println("          Save as ~/.gcal-organizer/credentials.json")
			failed++
		}

		// 4. token.json (OAuth token)
		tokenFile := filepath.Join(configDir, "token.json")
		if _, err := os.Stat(tokenFile); err == nil {
			// Check if token is valid
			f, err := os.Open(tokenFile)
			if err == nil {
				defer f.Close()
				var tok oauth2.Token
				if err := json.NewDecoder(f).Decode(&tok); err == nil {
					if tok.Expiry.After(time.Now()) || tok.RefreshToken != "" {
						fmt.Println("  ✅ OAuth token found (authenticated)")
						passed++
					} else {
						fmt.Println("  ⚠️  OAuth token exists but may be expired")
						fmt.Println("     Fix: Run 'gcal-organizer auth login' to re-authenticate")
						warned++
					}
				} else {
					fmt.Println("  ⚠️  OAuth token file is corrupted")
					fmt.Println("     Fix: Run 'gcal-organizer auth login' to re-authenticate")
					warned++
				}
			}
		} else {
			fmt.Println("  ❌ Not authenticated — no OAuth token found")
			fmt.Println("     Fix: Run 'gcal-organizer auth login'")
			failed++
		}

		// 5. GEMINI_API_KEY
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			// Try loading from .env
			apiKey = loadEnvValue(envFile, "GEMINI_API_KEY")
		}
		if apiKey != "" && apiKey != "your-gcp-api-key-here" {
			fmt.Printf("  ✅ GEMINI_API_KEY is set (%s****)\n", apiKey[:4])
			passed++
		} else if apiKey == "your-gcp-api-key-here" {
			fmt.Println("  ❌ GEMINI_API_KEY is still set to placeholder value")
			fmt.Println("     Fix: Get your API key from https://aistudio.google.com/app/apikey")
			fmt.Println("          Set it in ~/.gcal-organizer/.env")
			failed++
		} else {
			fmt.Println("  ❌ GEMINI_API_KEY is not set")
			fmt.Println("     Fix: Set GEMINI_API_KEY in ~/.gcal-organizer/.env or run 'gcal-organizer init'")
			failed++
		}

		// 6. Node.js
		if nodeOut, err := exec.Command("node", "--version").Output(); err == nil {
			version := strings.TrimSpace(string(nodeOut))
			fmt.Printf("  ✅ Node.js found (%s)\n", version)
			passed++
		} else {
			fmt.Println("  ⚠️  Node.js not found — Step 3 (task assignment) will be unavailable")
			fmt.Println("     Fix: Install Node.js 18+ from https://nodejs.org")
			warned++
		}

		// 7. Chrome profile
		chromePath := detectChromeProfile()
		if chromePath != "" {
			if _, err := os.Stat(chromePath); err == nil {
				fmt.Printf("  ✅ Chrome profile found (%s)\n", filepath.Base(chromePath))
				passed++
			} else {
				fmt.Printf("  ⚠️  Chrome profile not found at %s\n", chromePath)
				fmt.Println("     Fix: Set CHROME_PROFILE_PATH in ~/.gcal-organizer/.env")
				fmt.Println("          Find yours at chrome://version → 'Profile Path'")
				warned++
			}
		} else {
			fmt.Println("  ⚠️  Chrome profile path not detected")
			fmt.Println("     Fix: Set CHROME_PROFILE_PATH in ~/.gcal-organizer/.env")
			warned++
		}

		// 8. Service status
		if isServiceInstalled() {
			fmt.Println("  ✅ Hourly service is installed")
			passed++
		} else {
			fmt.Println("  ⚠️  Hourly service is not installed")
			fmt.Println("     Fix: Run 'gcal-organizer install' to set up the hourly service")
			warned++
		}

		// Summary
		fmt.Println()
		fmt.Println("───────────────────────────────────────────────────────────")
		fmt.Printf("  ✅ %d passed  ⚠️  %d warnings  ❌ %d failed\n", passed, warned, failed)
		if failed > 0 {
			fmt.Println()
			fmt.Println("  Run 'gcal-organizer init' to fix most issues.")
		} else if warned > 0 {
			fmt.Println()
			fmt.Println("  All critical checks passed! Warnings are informational.")
		} else {
			fmt.Println()
			fmt.Println("  🎉 Everything looks good!")
		}
		fmt.Println("═══════════════════════════════════════════════════════════")
		return nil
	},
}

// initCmd sets up the gcal-organizer configuration.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up gcal-organizer configuration",
	Long:  `Create the config directory and generate an environment file with your API keys.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		configDir := filepath.Join(home, ".gcal-organizer")
		envFile := filepath.Join(configDir, ".env")
		nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
		apiKey, _ := cmd.Flags().GetString("api-key")

		fmt.Println("🚀 gcal-organizer init")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println()

		// 1. Create config directory
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0700); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			fmt.Println("  ✅ Created ~/.gcal-organizer/")
		} else {
			fmt.Println("  ✅ Config directory already exists")
		}

		// 2. Generate .env file
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			// Get API key
			if apiKey == "" && !nonInteractive {
				fmt.Println()
				fmt.Print("  Enter your Gemini API key (from https://aistudio.google.com/app/apikey): ")
				fmt.Scanln(&apiKey)
			}
			if apiKey == "" {
				apiKey = "your-gcp-api-key-here"
			}

			// Detect Chrome profile
			chromePath := detectChromeProfile()

			// Write .env
			envContent := generateEnvFile(apiKey, chromePath)
			if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
				return fmt.Errorf("failed to write .env file: %w", err)
			}
			fmt.Println("  ✅ Created ~/.gcal-organizer/.env")
		} else {
			fmt.Println("  ✅ Environment file already exists (skipped)")
			// Check if API key is still placeholder
			existing := loadEnvValue(envFile, "GEMINI_API_KEY")
			if existing == "your-gcp-api-key-here" {
				fmt.Println("  ⚠️  GEMINI_API_KEY is still set to placeholder — edit ~/.gcal-organizer/.env")
			}
		}

		// 3. Check for credentials.json
		credFile := filepath.Join(configDir, "credentials.json")
		if _, err := os.Stat(credFile); os.IsNotExist(err) {
			fmt.Println()
			fmt.Println("  ⚠️  credentials.json not found")
			fmt.Println("     Download OAuth credentials from Google Cloud Console:")
			fmt.Println("     https://console.cloud.google.com/apis/credentials")
			fmt.Println("     Save as: ~/.gcal-organizer/credentials.json")
		} else {
			fmt.Println("  ✅ Google credentials found")
		}

		// 4. Summary
		fmt.Println()
		fmt.Println("───────────────────────────────────────────────────────────")
		fmt.Println("  Next steps:")
		if _, err := os.Stat(credFile); os.IsNotExist(err) {
			fmt.Println("  1. Download credentials.json (see above)")
			fmt.Println("  2. Run 'gcal-organizer auth login'")
		} else {
			tokenFile := filepath.Join(configDir, "token.json")
			if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
				fmt.Println("  1. Run 'gcal-organizer auth login'")
			} else {
				fmt.Println("  1. Run 'gcal-organizer run --dry-run' to test")
			}
		}
		fmt.Println("  2. Run 'gcal-organizer doctor' to verify setup")
		fmt.Println("═══════════════════════════════════════════════════════════")
		return nil
	},
}

// installCmd installs gcal-organizer as an hourly service.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install as an hourly background service",
	Long:  `Install gcal-organizer as an hourly service. Uses launchd on macOS and systemd on Linux.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		configDir := filepath.Join(home, ".gcal-organizer")

		// Check prerequisites
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			fmt.Println("⚠️  Config not found. Running 'gcal-organizer init' first...")
			fmt.Println()
			initCmd.RunE(cmd, args)
			fmt.Println()
		}

		// Find binary path
		binaryPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("could not determine binary path: %w", err)
		}

		fmt.Println("📦 gcal-organizer install")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println()

		switch runtime.GOOS {
		case "darwin":
			return installMacOS(home, binaryPath)
		case "linux":
			return installLinux(home, binaryPath)
		default:
			return fmt.Errorf("unsupported OS: %s (supported: darwin, linux)", runtime.GOOS)
		}
	},
}

// uninstallCmd removes the gcal-organizer service.
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the hourly background service",
	Long:  `Stop and remove the gcal-organizer service files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()

		fmt.Println("🗑️  gcal-organizer uninstall")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println()

		switch runtime.GOOS {
		case "darwin":
			return uninstallMacOS(home)
		case "linux":
			return uninstallLinux(home)
		default:
			return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
		}
	},
}

// --- Helper functions ---

// detectChromeProfile finds the default Chrome profile for the current OS.
func detectChromeProfile() string {
	home, _ := os.UserHomeDir()
	var baseDir string

	switch runtime.GOOS {
	case "darwin":
		baseDir = filepath.Join(home, "Library", "Application Support", "Google", "Chrome")
	case "linux":
		baseDir = filepath.Join(home, ".config", "google-chrome")
	default:
		return ""
	}

	// Try common profile names in order
	profiles := []string{"Default", "Profile 1", "Profile 2", "Profile 3"}
	for _, p := range profiles {
		path := filepath.Join(baseDir, p)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// loadEnvValue reads a single value from a .env file.
func loadEnvValue(envFile, key string) string {
	data, err := os.ReadFile(envFile)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if strings.TrimSpace(parts[0]) == key {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

// generateEnvFile creates the .env file content.
func generateEnvFile(apiKey, chromePath string) string {
	var b strings.Builder
	b.WriteString("# GCal Organizer Configuration\n")
	b.WriteString("# Generated by 'gcal-organizer init'\n\n")
	b.WriteString("# Required: GCP API Key for Gemini AI\n")
	b.WriteString(fmt.Sprintf("GEMINI_API_KEY=%s\n\n", apiKey))
	b.WriteString("# Required: Path to Google OAuth2 credentials\n")
	b.WriteString("GOOGLE_CREDENTIALS_FILE=~/.gcal-organizer/credentials.json\n\n")
	b.WriteString("# Optional: Master folder name in Google Drive\n")
	b.WriteString("GCAL_MASTER_FOLDER_NAME=Meeting Notes\n\n")
	b.WriteString("# Optional: Days to look back for calendar events\n")
	b.WriteString("GCAL_DAYS_TO_LOOK_BACK=8\n\n")
	b.WriteString("# Optional: Keywords to filter documents (comma-separated)\n")
	b.WriteString("GCAL_FILENAME_KEYWORDS=Notes,Meeting\n\n")
	b.WriteString("# Optional: Gemini model\n")
	b.WriteString("GEMINI_MODEL=gemini-2.0-flash\n")
	if chromePath != "" {
		b.WriteString(fmt.Sprintf("\n# Chrome profile for browser automation (auto-detected)\nCHROME_PROFILE_PATH=%s\n", chromePath))
	} else {
		b.WriteString("\n# Chrome profile for browser automation\n# Find yours at chrome://version → 'Profile Path'\n# CHROME_PROFILE_PATH=\n")
	}
	return b.String()
}

// isServiceInstalled checks if the hourly service is installed.
func isServiceInstalled() bool {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		plist := filepath.Join(home, "Library", "LaunchAgents", "com.jflowers.gcal-organizer.plist")
		_, err := os.Stat(plist)
		return err == nil
	case "linux":
		timer := filepath.Join(home, ".config", "systemd", "user", "gcal-organizer.timer")
		_, err := os.Stat(timer)
		return err == nil
	}
	return false
}

// --- macOS install/uninstall ---

func installMacOS(home, binaryPath string) error {
	logDir := filepath.Join(home, "Library", "Logs")
	logFile := filepath.Join(logDir, "gcal-organizer.log")
	plistDest := filepath.Join(home, "Library", "LaunchAgents", "com.jflowers.gcal-organizer.plist")
	wrapperDest := filepath.Join(home, ".local", "bin", "gcal-organizer-wrapper.sh")

	// Create wrapper script
	if err := os.MkdirAll(filepath.Dir(wrapperDest), 0755); err != nil {
		return fmt.Errorf("failed to create wrapper directory: %w", err)
	}

	wrapper := generateWrapper(binaryPath)
	if err := os.WriteFile(wrapperDest, []byte(wrapper), 0755); err != nil {
		return fmt.Errorf("failed to write wrapper script: %w", err)
	}
	fmt.Println("  ✅ Created wrapper script")

	// Create plist
	if err := os.MkdirAll(filepath.Dir(plistDest), 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	plist := generatePlist(wrapperDest, logFile, home, binaryPath)
	if err := os.WriteFile(plistDest, []byte(plist), 0644); err != nil {
		return fmt.Errorf("failed to write plist: %w", err)
	}
	fmt.Println("  ✅ Created LaunchAgent")

	// Load the service
	uid := fmt.Sprintf("%d", os.Getuid())
	exec.Command("launchctl", "bootout", "gui/"+uid, plistDest).Run() // ignore error
	if err := exec.Command("launchctl", "bootstrap", "gui/"+uid, plistDest).Run(); err != nil {
		return fmt.Errorf("failed to load LaunchAgent: %w\n   Fix: Try 'launchctl load %s'", err, plistDest)
	}
	fmt.Println("  ✅ Service loaded and running")

	fmt.Println()
	fmt.Println("───────────────────────────────────────────────────────────")
	fmt.Printf("  Logs:    %s\n", logFile)
	fmt.Println("  Status:  gcal-organizer doctor")
	fmt.Println("  Remove:  gcal-organizer uninstall")
	fmt.Println("═══════════════════════════════════════════════════════════")
	return nil
}

func uninstallMacOS(home string) error {
	plistDest := filepath.Join(home, "Library", "LaunchAgents", "com.jflowers.gcal-organizer.plist")
	wrapperDest := filepath.Join(home, ".local", "bin", "gcal-organizer-wrapper.sh")

	uid := fmt.Sprintf("%d", os.Getuid())
	exec.Command("launchctl", "bootout", "gui/"+uid, plistDest).Run()
	fmt.Println("  ✅ Service stopped")

	os.Remove(plistDest)
	fmt.Println("  ✅ Removed LaunchAgent")

	os.Remove(wrapperDest)
	fmt.Println("  ✅ Removed wrapper script")

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  Service fully removed.")
	fmt.Println("═══════════════════════════════════════════════════════════")
	return nil
}

// --- Linux install/uninstall ---

func installLinux(home, binaryPath string) error {
	systemdDir := filepath.Join(home, ".config", "systemd", "user")
	wrapperDest := filepath.Join(home, ".local", "bin", "gcal-organizer-wrapper.sh")

	// Create wrapper
	if err := os.MkdirAll(filepath.Dir(wrapperDest), 0755); err != nil {
		return fmt.Errorf("failed to create wrapper directory: %w", err)
	}
	wrapper := generateWrapper(binaryPath)
	if err := os.WriteFile(wrapperDest, []byte(wrapper), 0755); err != nil {
		return fmt.Errorf("failed to write wrapper script: %w", err)
	}
	fmt.Println("  ✅ Created wrapper script")

	// Create systemd directory
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	// Write service unit
	service := generateSystemdService(wrapperDest, home)
	if err := os.WriteFile(filepath.Join(systemdDir, "gcal-organizer.service"), []byte(service), 0644); err != nil {
		return fmt.Errorf("failed to write service unit: %w", err)
	}
	fmt.Println("  ✅ Created systemd service")

	// Write timer unit
	timer := generateSystemdTimer()
	if err := os.WriteFile(filepath.Join(systemdDir, "gcal-organizer.timer"), []byte(timer), 0644); err != nil {
		return fmt.Errorf("failed to write timer unit: %w", err)
	}
	fmt.Println("  ✅ Created systemd timer")

	// Enable and start
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	if err := exec.Command("systemctl", "--user", "enable", "--now", "gcal-organizer.timer").Run(); err != nil {
		return fmt.Errorf("failed to enable timer: %w\n   Fix: Check 'systemctl --user status gcal-organizer.timer'", err)
	}
	fmt.Println("  ✅ Timer enabled and started")

	fmt.Println()
	fmt.Println("───────────────────────────────────────────────────────────")
	fmt.Println("  Logs:    journalctl --user -u gcal-organizer.service")
	fmt.Println("  Status:  gcal-organizer doctor")
	fmt.Println("  Remove:  gcal-organizer uninstall")
	fmt.Println("═══════════════════════════════════════════════════════════")
	return nil
}

func uninstallLinux(home string) error {
	systemdDir := filepath.Join(home, ".config", "systemd", "user")
	wrapperDest := filepath.Join(home, ".local", "bin", "gcal-organizer-wrapper.sh")

	exec.Command("systemctl", "--user", "disable", "--now", "gcal-organizer.timer").Run()
	fmt.Println("  ✅ Timer stopped and disabled")

	os.Remove(filepath.Join(systemdDir, "gcal-organizer.service"))
	os.Remove(filepath.Join(systemdDir, "gcal-organizer.timer"))
	fmt.Println("  ✅ Removed systemd units")

	os.Remove(wrapperDest)
	fmt.Println("  ✅ Removed wrapper script")

	exec.Command("systemctl", "--user", "daemon-reload").Run()

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  Service fully removed.")
	fmt.Println("═══════════════════════════════════════════════════════════")
	return nil
}

// --- Embedded service templates ---

func generateWrapper(binaryPath string) string {
	return fmt.Sprintf(`#!/bin/bash
# gcal-organizer service wrapper
# Generated by 'gcal-organizer install'

set -euo pipefail

# Source env file if it exists
ENV_FILE="${HOME}/.gcal-organizer/.env"
if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE"
    set +a
fi

# Override days to look back for service mode (1 day)
export GCAL_DAYS_TO_LOOK_BACK=1

echo "$(date '+%%Y-%%m-%%d %%H:%%M:%%S') — Starting gcal-organizer run"
%s run
echo "$(date '+%%Y-%%m-%%d %%H:%%M:%%S') — Completed gcal-organizer run"
`, binaryPath)
}

func generatePlist(wrapperPath, logPath, home, binaryPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.jflowers.gcal-organizer</string>

    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>%s</string>
    </array>

    <key>StartInterval</key>
    <integer>3600</integer>

    <key>RunAtLoad</key>
    <true/>

    <key>StandardOutPath</key>
    <string>%s</string>

    <key>StandardErrorPath</key>
    <string>%s</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/opt/homebrew/bin</string>
        <key>HOME</key>
        <string>%s</string>
        <key>GCAL_ORGANIZER_BIN</key>
        <string>%s</string>
    </dict>
</dict>
</plist>
`, wrapperPath, logPath, logPath, home, binaryPath)
}

func generateSystemdService(wrapperPath, home string) string {
	return fmt.Sprintf(`[Unit]
Description=GCal Organizer - Meeting note organization
Documentation=https://github.com/jflowers/gcal-organizer

[Service]
Type=oneshot
ExecStart=/bin/bash %s
EnvironmentFile=-%s/.gcal-organizer/.env
Environment="HOME=%s"

[Install]
WantedBy=default.target
`, wrapperPath, home, home)
}

func generateSystemdTimer() string {
	return `[Unit]
Description=Run gcal-organizer hourly

[Timer]
OnCalendar=hourly
Persistent=true
RandomizedDelaySec=120

[Install]
WantedBy=timers.target
`
}
