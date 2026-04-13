package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/virtuallytd/claude-code-switcher/internal/profile"
)

// MergeSettings merges MCP servers from profile settings into ~/.claude.json
func MergeSettings(profilePath string) error {
	// Load profile settings
	profileSettings, err := profile.LoadSettings(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile settings: %w", err)
	}

	claudeStatePath := profile.GetClaudeStatePath()

	// Load ~/.claude.json
	data, err := os.ReadFile(claudeStatePath)
	if err != nil {
		return fmt.Errorf("failed to read ~/.claude.json: %w", err)
	}

	var claudeState map[string]interface{}
	if err := json.Unmarshal(data, &claudeState); err != nil {
		return fmt.Errorf("failed to parse ~/.claude.json: %w", err)
	}

	// Update top-level mcpServers key
	if profileSettings.MCPServers != nil {
		claudeState["mcpServers"] = profileSettings.MCPServers
	} else {
		delete(claudeState, "mcpServers")
	}

	// Write back
	output, err := json.MarshalIndent(claudeState, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ~/.claude.json: %w", err)
	}

	if err := os.WriteFile(claudeStatePath, output, 0600); err != nil {
		return fmt.Errorf("failed to write ~/.claude.json: %w", err)
	}

	return nil
}
