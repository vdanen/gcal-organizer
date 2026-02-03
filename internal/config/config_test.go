package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MasterFolderName != "Meeting Notes" {
		t.Errorf("expected MasterFolderName 'Meeting Notes', got %s", cfg.MasterFolderName)
	}

	if cfg.DaysToLookBack != 8 {
		t.Errorf("expected DaysToLookBack 8, got %d", cfg.DaysToLookBack)
	}

	if cfg.GeminiModel != "gemini-1.5-flash" {
		t.Errorf("expected GeminiModel 'gemini-1.5-flash', got %s", cfg.GeminiModel)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set up test environment
	os.Setenv("GCAL_MASTER_FOLDER_NAME", "Test Folder")
	os.Setenv("GCAL_DAYS_TO_LOOK_BACK", "14")
	os.Setenv("GEMINI_API_KEY", "test-api-key")
	defer func() {
		os.Unsetenv("GCAL_MASTER_FOLDER_NAME")
		os.Unsetenv("GCAL_DAYS_TO_LOOK_BACK")
		os.Unsetenv("GEMINI_API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.MasterFolderName != "Test Folder" {
		t.Errorf("expected MasterFolderName 'Test Folder', got %s", cfg.MasterFolderName)
	}

	if cfg.GeminiAPIKey != "test-api-key" {
		t.Errorf("expected GeminiAPIKey 'test-api-key', got %s", cfg.GeminiAPIKey)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		wantError bool
	}{
		{"empty key", "", true},
		{"valid key", "test-key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.GeminiAPIKey = tt.apiKey
			err := cfg.Validate()

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
