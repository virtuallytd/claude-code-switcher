package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build the container image",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Building %s image...\n", imageName)
			return runPodman("build", "-t", imageName, containerDir())
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
				if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
					continue
				}
				settingsPath := filepath.Join(profilesDir(), e.Name(), "settings.json")
				if _, err := os.Stat(settingsPath); err == nil {
					fmt.Printf("  %s\n", e.Name())
				}
			}
			return nil
		},
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <name>",
		Short: "Create a new profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			dir := filepath.Join(profilesDir(), name)

			if _, err := os.Stat(dir); err == nil {
				return fmt.Errorf("profile '%s' already exists at %s", name, dir)
			}

			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create %s: %w", dir, err)
			}

			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Auth type (claude/api/vertex): ")
			authType, _ := reader.ReadString('\n')
			authType = strings.TrimSpace(authType)

			switch authType {
			case "vertex":
				fmt.Print("Vertex project ID: ")
				projectID, _ := reader.ReadString('\n')
				projectID = strings.TrimSpace(projectID)

				fmt.Print("Vertex region [global]: ")
				region, _ := reader.ReadString('\n')
				region = strings.TrimSpace(region)
				if region == "" {
					region = "global"
				}

				envContent := fmt.Sprintf("CLAUDE_CODE_USE_VERTEX=1\nCLOUD_ML_REGION=%s\nANTHROPIC_VERTEX_PROJECT_ID=%s\n", region, projectID)
				os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0600)
			case "api":
				fmt.Print("Anthropic API key: ")
				apiKey, _ := reader.ReadString('\n')
				apiKey = strings.TrimSpace(apiKey)
				envContent := fmt.Sprintf("ANTHROPIC_API_KEY=%s\n", apiKey)
				os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0600)
			case "claude":
				fmt.Println("You'll be prompted to log in on first launch.")
			default:
				fmt.Println("You'll be prompted to log in on first launch.")
			}

			fmt.Print("Model [claude-sonnet-4-6]: ")
			model, _ := reader.ReadString('\n')
			model = strings.TrimSpace(model)
			if model == "" {
				model = "claude-sonnet-4-6"
			}

			settings := map[string]string{"model": model, "theme": "auto"}
			settingsJSON, _ := json.MarshalIndent(settings, "", "  ")
			os.WriteFile(filepath.Join(dir, "settings.json"), append(settingsJSON, '\n'), 0644)

			permissions := map[string]any{
				"permissions": map[string]any{
					"allow": []string{"WebSearch"},
				},
			}
			permJSON, _ := json.MarshalIndent(permissions, "", "  ")
			os.WriteFile(filepath.Join(dir, "settings.local.json"), append(permJSON, '\n'), 0644)

			fmt.Printf("\nProfile '%s' created at %s\n", name, dir)
			fmt.Println("To add MCP servers, create mcp.json in that directory.")
			fmt.Printf("Launch with: ccs %s /path/to/project\n", name)
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

			if err := exec.Command("podman", "image", "exists", imageName).Run(); err != nil {
				fmt.Printf("Image '%s' not found. Building...\n", imageName)
				if err := runPodman("build", "-t", imageName, containerDir()); err != nil {
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
				if parsingClaudeArgs || strings.HasPrefix(a, "-") {
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

			containerName := fmt.Sprintf("%s-%s-%s-%s", containerPrefix, profile, pathHash(projectPath), randSuffix())

			emptyMCP := filepath.Join(profileDir, ".empty-mcp.json")
			if _, err := os.Stat(emptyMCP); os.IsNotExist(err) {
				os.WriteFile(emptyMCP, []byte("{\"mcpServers\":{}}\n"), 0644)
			}

			authDir := filepath.Join(profileDir, ".auth")
			os.MkdirAll(authDir, 0700)

			keyringDir := filepath.Join(profileDir, ".keyring")
			os.MkdirAll(keyringDir, 0700)

			podmanArgs := []string{
				"run", "--rm", "-it",
				"--name", containerName,
				"-v", authDir + ":/root/.claude",
				"-v", keyringDir + ":/root/.local/share/keyrings",
				"-v", profileDir + ":/ccs-profile:ro",
				"-v", projectPath + ":" + projectPath,
				"-w", projectPath,
			}

			if _, err := os.Stat(envFile); err == nil {
				podmanArgs = append(podmanArgs, "--env-file", envFile)
			}

			mcpFile := filepath.Join(profileDir, "mcp.json")
			if _, err := os.Stat(mcpFile); err == nil {
				dir := filepath.Dir(projectPath)
				for dir != "/" && dir != "." {
					parentMCP := filepath.Join(dir, ".mcp.json")
					if _, err := os.Stat(parentMCP); err == nil {
						podmanArgs = append(podmanArgs, "-v", emptyMCP+":"+parentMCP+":ro")
					}
					dir = filepath.Dir(dir)
				}
			}

			if envFileContains(envFile, "CLAUDE_CODE_USE_VERTEX") {
				home, _ := os.UserHomeDir()
				gcloudDir := filepath.Join(home, ".config", "gcloud")
				if _, err := os.Stat(gcloudDir); err == nil {
					podmanArgs = append(podmanArgs, "-v", gcloudDir+":/root/.config/gcloud")
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
