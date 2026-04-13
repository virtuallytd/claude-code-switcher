package cmd

import (
	"fmt"

	"github.com/virtuallytd/claude-code-switcher/internal/profile"
)

// Current shows the currently active profile
func Current() error {
	current, err := profile.GetCurrentProfile()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	if current == "" {
		fmt.Println("No profile currently active")
		return nil
	}

	// Verify the profile still exists
	profiles, err := profile.Discover()
	if err != nil {
		return err
	}

	found := false
	var p profile.Profile
	for _, prof := range profiles {
		if prof.Name == current {
			found = true
			p = prof
			break
		}
	}

	if !found {
		fmt.Printf("Profile '%s' (no longer exists)\n", current)
		return nil
	}

	authType := "standard"
	if p.IsVertex {
		authType = "vertex"
	}

	fmt.Printf("%s (%s)\n", current, authType)
	return nil
}
