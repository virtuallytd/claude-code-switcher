package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	imageName       = "claude-code"
	containerPrefix = "cc"
)

func profilesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ccs_profiles")
}

func binaryDir() string {
	exe, err := os.Executable()
	if err != nil {
		wd, _ := os.Getwd()
		return wd
	}
	return filepath.Dir(exe)
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

func main() {
	root := &cobra.Command{
		Use:   "ccs <profile> [path] [-- claude-args...]",
		Short: "Claude Code Switcher — run Claude Code with isolated Podman profiles",
		Example: `  ccs work ~/Projects/work/my-repo
  ccs personal ~/Projects/personal/my-app
  ccs work ~/Projects/work/my-repo -- --resume
  ccs build
  ccs stop work`,
	}

	root.AddCommand(buildCmd())
	root.AddCommand(stopCmd())
	root.AddCommand(statusCmd())
	root.AddCommand(profilesCmd())
	root.AddCommand(runCmd())

	// Route unknown subcommands to "run" so `ccs work` works like `ccs run work`
	knownCmds := map[string]bool{"build": true, "stop": true, "status": true, "profiles": true, "run": true, "help": true, "completion": true}
	if len(os.Args) > 1 && !knownCmds[os.Args[1]] && !strings.HasPrefix(os.Args[1], "-") {
		os.Args = append([]string{os.Args[0], "run"}, os.Args[1:]...)
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build the container image",
		RunE: func(cmd *cobra.Command, args []string) error {
			uid := os.Getuid()
			gid := os.Getgid()
			fmt.Printf("Building %s image (UID=%d, GID=%d)...\n", imageName, uid, gid)
			return runPodman(
				"build",
				"--build-arg", fmt.Sprintf("USER_UID=%d", uid),
				"--build-arg", fmt.Sprintf("USER_GID=%d", gid),
				"-t", imageName,
				binaryDir(),
			)
		},
	}
}

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop [profile]",
		Short: "Stop running container(s)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := containerPrefix + "-"
			label := "all claude-code"
			if len(args) == 1 {
				filter = containerPrefix + "-" + args[0] + "-"
				label = args[0]
			}

			containers, _ := podmanOutput("ps", "-q", "--filter", "name="+filter)
			if containers == "" {
				fmt.Printf("No running %s containers.\n", label)
				return nil
			}

			fmt.Printf("Stopping %s containers...\n", label)
			ids := strings.Fields(containers)
			podmanArgs := append([]string{"stop"}, ids...)
			return runPodman(podmanArgs...)
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show running containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPodman("ps",
				"--filter", "name="+containerPrefix+"-",
				"--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
		},
	}
}

func profilesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "profiles",
		Short: "List available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := os.ReadDir(profilesDir())
			if err != nil {
				return fmt.Errorf("cannot read %s: %w", profilesDir(), err)
			}

			fmt.Println("Available profiles:")
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				envPath := filepath.Join(profilesDir(), e.Name(), ".env")
				if _, err := os.Stat(envPath); err == nil {
					fmt.Printf("  %s\n", e.Name())
				}
			}
			return nil
		},
	}
}

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "run <profile> [path] [-- claude-args...]",
		Short:              "Launch Claude Code with the given profile",
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := args[0]
			rest := args[1:]

			profileDir := filepath.Join(profilesDir(), profile)
			if _, err := os.Stat(profileDir); os.IsNotExist(err) {
				return fmt.Errorf("profile '%s' not found in %s/", profile, profilesDir())
			}

			envFile := filepath.Join(profileDir, ".env")
			if _, err := os.Stat(envFile); os.IsNotExist(err) {
				return fmt.Errorf("%s not found", envFile)
			}

			if err := exec.Command("podman", "image", "exists", imageName).Run(); err != nil {
				fmt.Printf("Image '%s' not found. Building...\n", imageName)
				uid := os.Getuid()
				gid := os.Getgid()
				if err := runPodman("build",
					"--build-arg", fmt.Sprintf("USER_UID=%d", uid),
					"--build-arg", fmt.Sprintf("USER_GID=%d", gid),
					"-t", imageName, binaryDir()); err != nil {
					return err
				}
			}

			var projectPath string
			var claudeArgs []string
			parsingClaudeArgs := false

			for _, a := range rest {
				if a == "--" {
					parsingClaudeArgs = true
					continue
				}
				if parsingClaudeArgs {
					claudeArgs = append(claudeArgs, a)
				} else if projectPath == "" {
					projectPath = a
				} else {
					claudeArgs = append(claudeArgs, a)
				}
			}

			if projectPath == "" {
				projectPath, _ = os.Getwd()
			}

			if abs, err := filepath.Abs(projectPath); err == nil {
				projectPath = abs
			}

			containerName := fmt.Sprintf("%s-%s-%s", containerPrefix, profile, pathHash(projectPath))

			_ = exec.Command("podman", "rm", "-f", containerName).Run()

			podmanArgs := []string{
				"run", "--rm", "-it",
				"--name", containerName,
				"--env-file", envFile,
				"-v", filepath.Join(profileDir, "settings.json") + ":/home/claude/.claude/settings.json:ro",
				"-v", filepath.Join(profileDir, "settings.local.json") + ":/home/claude/.claude/settings.local.json:ro",
				"-v", projectPath + ":" + projectPath,
				"-w", projectPath,
			}

			mcpFile := filepath.Join(profileDir, "mcp.json")
			if _, err := os.Stat(mcpFile); err == nil {
				podmanArgs = append(podmanArgs, "-v", mcpFile+":"+projectPath+"/.mcp.json:ro")
			}

			if envFileContains(envFile, "CLAUDE_CODE_USE_VERTEX") {
				home, _ := os.UserHomeDir()
				gcloudDir := filepath.Join(home, ".config", "gcloud")
				if _, err := os.Stat(gcloudDir); err == nil {
					podmanArgs = append(podmanArgs, "-v", gcloudDir+":/home/claude/.config/gcloud:ro")
				}
			}

			needsHostNet := fileContainsString(mcpFile, "localhost") ||
				fileContainsString(filepath.Join(profileDir, "settings.local.json"), "localhost")
			if needsHostNet {
				podmanArgs = append(podmanArgs, "--network", "host")
			}

			podmanArgs = append(podmanArgs, imageName)
			podmanArgs = append(podmanArgs, claudeArgs...)

			return runPodman(podmanArgs...)
		},
	}
}
