package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected EnvVars
		wantErr  bool
	}{
		{
			name: "valid vertex config",
			content: `export CLAUDE_CODE_USE_VERTEX=1
export CLOUD_ML_REGION=us-east5
export ANTHROPIC_VERTEX_PROJECT_ID=my-project-123`,
			expected: EnvVars{
				UseVertex: true,
				Region:    "us-east5",
				ProjectID: "my-project-123",
			},
			wantErr: false,
		},
		{
			name: "quoted values",
			content: `export CLAUDE_CODE_USE_VERTEX="1"
export CLOUD_ML_REGION="us-central1"
export ANTHROPIC_VERTEX_PROJECT_ID='my-project'`,
			expected: EnvVars{
				UseVertex: true,
				Region:    "us-central1",
				ProjectID: "my-project",
			},
			wantErr: false,
		},
		{
			name: "mixed format with comments",
			content: `# Vertex AI configuration
export CLAUDE_CODE_USE_VERTEX=true

# Region setting
export CLOUD_ML_REGION=eu-west1
export ANTHROPIC_VERTEX_PROJECT_ID=test-project`,
			expected: EnvVars{
				UseVertex: true,
				Region:    "eu-west1",
				ProjectID: "test-project",
			},
			wantErr: false,
		},
		{
			name: "extra whitespace",
			content: `  export  CLAUDE_CODE_USE_VERTEX = 1
export CLOUD_ML_REGION  =  us-west2
export ANTHROPIC_VERTEX_PROJECT_ID=my-project  `,
			expected: EnvVars{
				UseVertex: true,
				Region:    "us-west2",
				ProjectID: "my-project",
			},
			wantErr: false,
		},
		{
			name: "partial config",
			content: `export CLOUD_ML_REGION=us-east5
export ANTHROPIC_VERTEX_PROJECT_ID=my-project`,
			expected: EnvVars{
				UseVertex: false,
				Region:    "us-east5",
				ProjectID: "my-project",
			},
			wantErr: false,
		},
		{
			name:    "empty file",
			content: "",
			expected: EnvVars{
				UseVertex: false,
				Region:    "",
				ProjectID: "",
			},
			wantErr: false,
		},
		{
			name: "non-export lines ignored",
			content: `CLAUDE_CODE_USE_VERTEX=1
export CLOUD_ML_REGION=us-east5
ANTHROPIC_VERTEX_PROJECT_ID=ignored
export ANTHROPIC_VERTEX_PROJECT_ID=my-project`,
			expected: EnvVars{
				UseVertex: false,
				Region:    "us-east5",
				ProjectID: "my-project",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and env.zsh file
			tmpDir := t.TempDir()
			envPath := filepath.Join(tmpDir, "env.zsh")

			if err := os.WriteFile(envPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test LoadEnvVars
			result, err := LoadEnvVars(tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadEnvVars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.UseVertex != tt.expected.UseVertex {
				t.Errorf("UseVertex = %v, want %v", result.UseVertex, tt.expected.UseVertex)
			}
			if result.Region != tt.expected.Region {
				t.Errorf("Region = %q, want %q", result.Region, tt.expected.Region)
			}
			if result.ProjectID != tt.expected.ProjectID {
				t.Errorf("ProjectID = %q, want %q", result.ProjectID, tt.expected.ProjectID)
			}
		})
	}
}

func TestLoadEnvVarsFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadEnvVars(tmpDir)
	if err == nil {
		t.Error("Expected error when env.zsh doesn't exist, got nil")
	}
}

func TestLoadSettings(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid settings",
			content: `{
  "mcpServers": {
    "test-server": {
      "type": "sse",
      "url": "http://localhost:3000/sse"
    }
  }
}`,
			wantErr: false,
		},
		{
			name:    "empty settings",
			content: `{}`,
			wantErr: false,
		},
		{
			name: "settings with other fields",
			content: `{
  "mcpServers": {
    "server1": {
      "type": "sse",
      "url": "http://localhost:3000/sse"
    }
  },
  "otherField": "ignored"
}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			content: `{invalid json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			settingsPath := filepath.Join(tmpDir, "settings.json")

			if err := os.WriteFile(settingsPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result, err := LoadSettings(tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("LoadSettings() returned nil result without error")
			}
		})
	}
}

func TestLoadSettingsFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadSettings(tmpDir)
	if err == nil {
		t.Error("Expected error when settings.json doesn't exist, got nil")
	}
}

func TestGetProfilesDir(t *testing.T) {
	dir := GetProfilesDir()
	if dir == "" {
		t.Error("GetProfilesDir() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("GetProfilesDir() = %q is not absolute path", dir)
	}
	// Should end with .claude/profiles
	if filepath.Base(dir) != "profiles" {
		t.Errorf("GetProfilesDir() = %q doesn't end with 'profiles'", dir)
	}
}

func TestGetClaudeStatePath(t *testing.T) {
	path := GetClaudeStatePath()
	if path == "" {
		t.Error("GetClaudeStatePath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("GetClaudeStatePath() = %q is not absolute path", path)
	}
	// Should end with .claude.json
	if filepath.Base(path) != ".claude.json" {
		t.Errorf("GetClaudeStatePath() = %q doesn't end with '.claude.json'", path)
	}
}

func TestSetAndGetCurrentProfile(t *testing.T) {
	// Set up temp home directory
	tmpHome := t.TempDir()
	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	// Temporarily override home dir for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	profileName := "test-profile"

	// Test SetCurrentProfile
	if err := SetCurrentProfile(profileName); err != nil {
		t.Fatalf("SetCurrentProfile() error = %v", err)
	}

	// Test GetCurrentProfile
	result, err := GetCurrentProfile()
	if err != nil {
		t.Fatalf("GetCurrentProfile() error = %v", err)
	}

	if result != profileName {
		t.Errorf("GetCurrentProfile() = %q, want %q", result, profileName)
	}
}

func TestGetCurrentProfileNotFound(t *testing.T) {
	// Set up temp home directory without state file
	tmpHome := t.TempDir()
	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	result, err := GetCurrentProfile()
	if err != nil {
		t.Errorf("GetCurrentProfile() unexpected error = %v", err)
	}
	if result != "" {
		t.Errorf("GetCurrentProfile() = %q, want empty string when file doesn't exist", result)
	}
}
