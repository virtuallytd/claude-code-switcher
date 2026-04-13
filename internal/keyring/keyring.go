package keyring

import (
	"fmt"
	"os/user"
	"runtime"

	"github.com/99designs/keyring"
)

const (
	serviceName       = "Claude Code"
	profileServiceFmt = "Claude Code - %s"
)

// getKeyring returns a platform-specific keyring
func getKeyring() (keyring.Keyring, error) {
	var backend keyring.BackendType

	switch runtime.GOOS {
	case "darwin":
		backend = keyring.KeychainBackend
	case "linux":
		backend = keyring.SecretServiceBackend
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	config := keyring.Config{
		ServiceName:              serviceName,
		AllowedBackends:          []keyring.BackendType{backend},
		KeychainTrustApplication: true,
	}

	kr, err := keyring.Open(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	return kr, nil
}

// activeTokenKey returns the keychain account name Claude Code uses for the active token.
// Claude Code stores the token under the OS username, not the service name.
func activeTokenKey() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return u.Username, nil
}

// GetActiveToken retrieves the active Claude Code session token
func GetActiveToken() (string, error) {
	kr, err := getKeyring()
	if err != nil {
		return "", err
	}

	key, err := activeTokenKey()
	if err != nil {
		return "", err
	}

	item, err := kr.Get(key)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", fmt.Errorf("no active Claude session found. Run 'claude login' first")
		}
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	return string(item.Data), nil
}

// SetActiveToken sets the active Claude Code session token
func SetActiveToken(token string) error {
	kr, err := getKeyring()
	if err != nil {
		return err
	}

	key, err := activeTokenKey()
	if err != nil {
		return err
	}

	item := keyring.Item{
		Key:  key,
		Data: []byte(token),
	}

	if err := kr.Set(item); err != nil {
		return fmt.Errorf("failed to set token: %w", err)
	}

	return nil
}

// GetProfileToken retrieves a saved profile token
func GetProfileToken(profileName string) (string, error) {
	kr, err := getKeyring()
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf(profileServiceFmt, profileName)
	item, err := kr.Get(key)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", fmt.Errorf("no saved token for profile '%s'. Run: claude login && ccs save %s", profileName, profileName)
		}
		return "", fmt.Errorf("failed to get profile token: %w", err)
	}

	return string(item.Data), nil
}

// SaveProfileToken saves the current active token as a named profile
func SaveProfileToken(profileName string, token string) error {
	kr, err := getKeyring()
	if err != nil {
		return err
	}

	key := fmt.Sprintf(profileServiceFmt, profileName)
	item := keyring.Item{
		Key:  key,
		Data: []byte(token),
	}

	if err := kr.Set(item); err != nil {
		return fmt.Errorf("failed to save profile token: %w", err)
	}

	return nil
}
