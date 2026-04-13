package cmd

import (
	"fmt"
	"os"

	"github.com/virtuallytd/claude-code-switcher/internal/keyring"
)

// Save saves the current active session token as a named profile token
func Save(profileName string) error {
	token, err := keyring.GetActiveToken()
	if err != nil {
		return fmt.Errorf("failed to get active token: %w", err)
	}

	if err := keyring.SaveProfileToken(profileName, token); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Saved session token for profile '%s'\n", profileName)
	return nil
}
