package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jflowers/gcal-organizer/internal/auth"
	"github.com/jflowers/gcal-organizer/internal/config"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// config command group
// ---------------------------------------------------------------------------

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `View and manage gcal-organizer configuration.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the merged configuration from environment variables and config file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Println("📋 Current Configuration:")
		fmt.Println("───────────────────────────────────────────────────────────")
		fmt.Printf("   Master Folder:     %s\n", cfg.MasterFolderName)
		fmt.Printf("   Days to Look Back: %d\n", cfg.DaysToLookBack)
		fmt.Printf("   Filename Pattern:  %s\n", cfg.FilenamePattern)
		fmt.Printf("   Gemini Model:      %s\n", cfg.GeminiModel)
		fmt.Printf("   Gemini API Key:    %s\n", maskSecret(cfg.GeminiAPIKey))
		fmt.Printf("   Credentials File:  %s\n", cfg.CredentialsFile)
		fmt.Printf("   Token File:        %s\n", cfg.TokenFile)
		fmt.Println("───────────────────────────────────────────────────────────")

		if _, err := os.Stat(cfg.CredentialsFile); os.IsNotExist(err) {
			fmt.Println("⚠️  OAuth credentials not found!")
			fmt.Printf("   Download from Google Cloud Console and save to:\n   %s\n", cfg.CredentialsFile)
		} else {
			fmt.Println("✅ OAuth credentials file found")
		}

		if _, err := os.Stat(cfg.TokenFile); os.IsNotExist(err) {
			fmt.Println("⚠️  Not authenticated - run 'gcal-organizer auth login'")
		} else {
			fmt.Println("✅ OAuth token found (authenticated)")
		}

		return nil
	},
}

// ---------------------------------------------------------------------------
// auth command group
// ---------------------------------------------------------------------------

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication management",
	Long:  `Manage OAuth authentication for Google Workspace APIs.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Google Workspace",
	Long:  `Start the OAuth2 flow to authenticate with Google Workspace APIs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Println("🔐 Starting OAuth2 login flow...")
		fmt.Println("")

		oauthClient, err := auth.NewOAuthClient(cfg.CredentialsFile, cfg.TokenFile)
		if err != nil {
			return fmt.Errorf("failed to create OAuth client: %w\n\nTo set up OAuth:\n1. Go to https://console.cloud.google.com\n2. Create OAuth 2.0 credentials (Desktop app)\n3. Download and save to: %s\n\nRun 'gcal-organizer doctor' for full diagnostics", err, cfg.CredentialsFile)
		}

		_, err = oauthClient.GetClient(ctx)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		fmt.Println("")
		fmt.Println("✅ Authentication successful!")
		fmt.Printf("   Token saved to: %s\n", cfg.TokenFile)
		fmt.Println("")
		fmt.Println("You can now run: gcal-organizer run --dry-run")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Check if OAuth authentication is configured and valid.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Println("🔐 Authentication Status:")
		fmt.Println("───────────────────────────────────────────────────────────")

		if _, err := os.Stat(cfg.CredentialsFile); os.IsNotExist(err) {
			fmt.Println("❌ OAuth credentials file NOT found")
			fmt.Printf("   Expected: %s\n", cfg.CredentialsFile)
			fmt.Println("")
			fmt.Println("To fix: Download OAuth credentials from Google Cloud Console")
			return nil
		}
		fmt.Println("✅ OAuth credentials file found")

		if _, err := os.Stat(cfg.TokenFile); os.IsNotExist(err) {
			fmt.Println("❌ Not authenticated")
			fmt.Println("")
			fmt.Println("Run: gcal-organizer auth login")
			return nil
		}

		oauthClient, err := auth.NewOAuthClient(cfg.CredentialsFile, cfg.TokenFile)
		if err != nil {
			fmt.Printf("❌ Error loading credentials: %v\n", err)
			return nil
		}

		if oauthClient.IsAuthenticated() {
			fmt.Println("✅ Authenticated and token valid")
		} else {
			fmt.Println("⚠️  Token expired - run 'gcal-organizer auth login' to refresh")
		}

		fmt.Println("")
		if cfg.GeminiAPIKey != "" {
			fmt.Println("✅ Gemini API key configured")
		} else {
			fmt.Println("❌ Gemini API key NOT set")
			fmt.Println("   Set GEMINI_API_KEY environment variable")
		}

		return nil
	},
}

// maskSecret returns a partially-redacted version of a secret string for display.
func maskSecret(s string) string {
	if len(s) == 0 {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
