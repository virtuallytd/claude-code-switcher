package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMergeSettings(t *testing.T) {
	// Set up temp home directory
	tmpHome := t.TempDir()
	claudeStatePath := filepath.Join(tmpHome, ".claude.json")

	// Create initial claude state
	initialState := map[string]interface{}{
		"existingKey": "existingValue",
		"mcpServers": map[string]interface{}{
			"old-server": map[string]interface{}{
				"type": "sse",
				"url":  "http://localhost:9999/sse",
			},
		},
		"otherSetting": 123,
	}

	stateJSON, err := json.MarshalIndent(initialState, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal initial state: %v", err)
	}

	if err := os.WriteFile(claudeStatePath, stateJSON, 0600); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	// Create profile settings
	profileDir := t.TempDir()
	profileSettings := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"new-server": map[string]interface{}{
				"type": "sse",
				"url":  "http://localhost:3000/sse",
			},
		},
	}

	profileJSON, err := json.MarshalIndent(profileSettings, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal profile settings: %v", err)
	}

	settingsPath := filepath.Join(profileDir, "settings.json")
	if err := os.WriteFile(settingsPath, profileJSON, 0644); err != nil {
		t.Fatalf("Failed to write profile settings: %v", err)
	}

	// Override home for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Test MergeSettings
	if err := MergeSettings(profileDir); err != nil {
		t.Fatalf("MergeSettings() error = %v", err)
	}

	// Read result
	resultData, err := os.ReadFile(claudeStatePath)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resultData, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify existing keys preserved
	if result["existingKey"] != "existingValue" {
		t.Errorf("existingKey = %v, want existingValue", result["existingKey"])
	}

	if result["otherSetting"].(float64) != 123 {
		t.Errorf("otherSetting = %v, want 123", result["otherSetting"])
	}

	// Verify mcpServers replaced
	mcpServers, ok := result["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("mcpServers is not a map")
	}

	if _, exists := mcpServers["old-server"]; exists {
		t.Error("old-server should be removed")
	}

	if _, exists := mcpServers["new-server"]; !exists {
		t.Error("new-server should be present")
	}
}

func TestMergeSettingsEmptyMCPServers(t *testing.T) {
	// Set up temp home directory
	tmpHome := t.TempDir()
	claudeStatePath := filepath.Join(tmpHome, ".claude.json")

	// Create initial claude state with mcpServers
	initialState := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"old-server": map[string]interface{}{
				"type": "sse",
				"url":  "http://localhost:9999/sse",
			},
		},
	}

	stateJSON, err := json.MarshalIndent(initialState, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal initial state: %v", err)
	}

	if err := os.WriteFile(claudeStatePath, stateJSON, 0600); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	// Create profile with no mcpServers
	profileDir := t.TempDir()
	profileSettings := map[string]interface{}{}

	profileJSON, err := json.MarshalIndent(profileSettings, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal profile settings: %v", err)
	}

	settingsPath := filepath.Join(profileDir, "settings.json")
	if err := os.WriteFile(settingsPath, profileJSON, 0644); err != nil {
		t.Fatalf("Failed to write profile settings: %v", err)
	}

	// Override home for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Test MergeSettings
	if err := MergeSettings(profileDir); err != nil {
		t.Fatalf("MergeSettings() error = %v", err)
	}

	// Read result
	resultData, err := os.ReadFile(claudeStatePath)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resultData, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify mcpServers removed
	if _, exists := result["mcpServers"]; exists {
		t.Error("mcpServers should be removed when profile has none")
	}
}

func TestMergeSettingsInvalidProfilePath(t *testing.T) {
	err := MergeSettings("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for nonexistent profile path, got nil")
	}
}
