package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	imageName       = "claude-code"
	containerPrefix = "cc"
)

func profilesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ccs_profiles")
}

func containerDir() string {
	exe, err := os.Executable()
	if err != nil {
		wd, _ := os.Getwd()
		return wd
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return filepath.Dir(exe)
	}
	dir := filepath.Dir(resolved)

	// Check for container/ subdirectory (source layout)
	sub := filepath.Join(dir, "container")
	if _, err := os.Stat(filepath.Join(sub, "Containerfile")); err == nil {
		return sub
	}

	// Homebrew: binary is in <prefix>/bin/, container files are in <prefix>/container/
	if filepath.Base(dir) == "bin" {
		parent := filepath.Dir(dir)
		sub := filepath.Join(parent, "container")
		if _, err := os.Stat(filepath.Join(sub, "Containerfile")); err == nil {
			return sub
		}
	}

	return dir
}

func runPodman(args ...string) error {
	cmd := exec.Command("podman", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func podmanOutput(args ...string) (string, error) {
	out, err := exec.Command("podman", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func envFileContains(path, key string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, key+"=") {
			val := strings.TrimPrefix(line, key+"=")
			return val == "1" || val == "true"
		}
	}
	return false
}

func fileContainsString(path, substr string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), substr)
}

func pathHash(p string) string {
	h := sha256.Sum256([]byte(p))
	return fmt.Sprintf("%x", h[:4])
}

func randSuffix() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}
