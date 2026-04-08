package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/adavis/ccs/internal/profile"
)

// MergeSettings merges MCP servers from profile settings into active settings.json
func MergeSettings(profilePath string) error {
	// Load profile settings
	profileSettings, err := profile.LoadSettings(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile settings: %w", err)
	}

	activeSettingsPath := profile.GetActiveSettingsPath()

	// Load active settings (or create empty if doesn't exist)
	var activeSettings map[string]interface{}
	data, err := os.ReadFile(activeSettingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			activeSettings = make(map[string]interface{})
		} else {
			return fmt.Errorf("failed to read active settings: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &activeSettings); err != nil {
			return fmt.Errorf("failed to parse active settings: %w", err)
		}
	}

	// Merge mcpServers
	if profileSettings.MCPServers != nil {
		activeSettings["mcpServers"] = profileSettings.MCPServers
	} else {
		// Clear mcpServers if profile has none
		delete(activeSettings, "mcpServers")
	}

	// Write back to active settings
	output, err := json.MarshalIndent(activeSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(activeSettingsPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write active settings: %w", err)
	}

	return nil
}
