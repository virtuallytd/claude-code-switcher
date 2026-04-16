package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Profile represents a Claude Code profile
type Profile struct {
	Name     string
	Path     string
	IsVertex bool // true if has env.zsh (Vertex AI auth)
}

// Settings represents Claude Code settings.json structure
type Settings struct {
	MCPServers map[string]interface{} `json:"mcpServers,omitempty"`
}

// EnvVars represents environment variables from env.zsh
type EnvVars struct {
	UseVertex bool
	Region    string
	ProjectID string
}

// GetProfilesDir returns the profiles directory path
func GetProfilesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "profiles")
}

// GetClaudeStatePath returns the path to ~/.claude.json (Claude Code's main config)
func GetClaudeStatePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude.json")
}

// GetActiveSettingsPath returns the path to active settings.json
func GetActiveSettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "settings.json")
}

const firstRunMessage = `Welcome to ccs! To get started, create a profile directory:

  mkdir -p ~/.claude/profiles/<profile-name>

Then add a settings.json with your MCP server configuration:

  ~/.claude/profiles/<profile-name>/settings.json

See https://github.com/virtuallytd/claude-code-switcher#profiles for details.`

// Discover finds all available profiles
func Discover() ([]Profile, error) {
	profilesDir := GetProfilesDir()

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(profilesDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create profiles directory: %w", err)
			}
			return nil, fmt.Errorf("%s", firstRunMessage)
		}
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var profiles []Profile
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		profilePath := filepath.Join(profilesDir, name)

		// Check if settings.json exists (required)
		settingsPath := filepath.Join(profilePath, "settings.json")
		if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
			continue
		}

		// Check if env.zsh exists (indicates Vertex AI profile)
		envPath := filepath.Join(profilePath, "env.zsh")
		isVertex := false
		if _, err := os.Stat(envPath); err == nil {
			isVertex = true
		}

		profiles = append(profiles, Profile{
			Name:     name,
			Path:     profilePath,
			IsVertex: isVertex,
		})
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("%s", firstRunMessage)
	}

	return profiles, nil
}

// LoadSettings loads the settings.json for a profile
func LoadSettings(profilePath string) (*Settings, error) {
	settingsPath := filepath.Join(profilePath, "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings.json: %w", err)
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings.json: %w", err)
	}

	return &settings, nil
}

// LoadEnvVars loads environment variables from env.zsh
func LoadEnvVars(profilePath string) (*EnvVars, error) {
	envPath := filepath.Join(profilePath, "env.zsh")

	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read env.zsh: %w", err)
	}

	vars := &EnvVars{}
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "export ") {
			continue
		}

		// Remove "export " prefix
		line = strings.TrimPrefix(line, "export ")

		// Split on first =
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch key {
		case "CLAUDE_CODE_USE_VERTEX":
			vars.UseVertex = value == "1" || value == "true"
		case "CLOUD_ML_REGION":
			vars.Region = value
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			vars.ProjectID = value
		}
	}

	return vars, nil
}

// GetCurrentProfile reads the current profile from a state file
func GetCurrentProfile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	statePath := filepath.Join(home, ".claude", ".current-profile")
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No current profile set
		}
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// SetCurrentProfile saves the current profile to a state file
func SetCurrentProfile(name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	statePath := filepath.Join(home, ".claude", ".current-profile")
	return os.WriteFile(statePath, []byte(name), 0644)
}
