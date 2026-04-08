package cmd

import (
	"fmt"

	"github.com/adavis/ccs/internal/profile"
)

// Reload re-applies the currently active profile's settings and auth
func Reload() error {
	current, err := profile.GetCurrentProfile()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	if current == "" {
		return fmt.Errorf("no profile currently active — use 'ccs switch' first")
	}

	profiles, err := profile.Discover()
	if err != nil {
		return err
	}

	for _, p := range profiles {
		if p.Name == current {
			if err := generateShellCommands(p); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("current profile '%s' no longer exists in ~/.claude/profiles/", current)
}
