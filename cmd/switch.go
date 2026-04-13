package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/adavis/ccs/internal/config"
	"github.com/adavis/ccs/internal/keyring"
	"github.com/adavis/ccs/internal/profile"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	vertexBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Render("●")

	standardBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Render("○")
)

type model struct {
	profiles       []profile.Profile
	cursor         int
	selected       int
	currentProfile string
	err            error
}

func initialModel() (model, error) {
	profiles, err := profile.Discover()
	if err != nil {
		return model{}, err
	}

	currentProfile, _ := profile.GetCurrentProfile()

	cursor := 0
	for i, p := range profiles {
		if p.Name == currentProfile {
			cursor = i
			break
		}
	}

	return model{
		profiles:       profiles,
		cursor:         cursor,
		selected:       -1,
		currentProfile: currentProfile,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}

		case "enter":
			m.selected = m.cursor
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Claude Code Switcher"))
	b.WriteString("\n\n")

	for i, p := range m.profiles {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render("❯")
		}

		// Use filled dot for active profile, empty dot for inactive
		badge := standardBadge // ○
		if p.Name == m.currentProfile {
			badge = vertexBadge // ●
		}

		authType := "standard"
		if p.IsVertex {
			authType = "vertex"
		}

		name := p.Name
		if m.cursor == i {
			name = selectedStyle.Render(name)
		}

		b.WriteString(fmt.Sprintf("%s %s %s %s\n", cursor, badge, name, helpStyle.Render(fmt.Sprintf("(%s)", authType))))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/k up • ↓/j down • enter select • q/esc quit"))
	b.WriteString("\n")

	return b.String()
}

// Switch runs the interactive profile switcher
func Switch() error {
	m, err := initialModel()
	if err != nil {
		return err
	}

	// Render TUI to stderr so only shell commands go to stdout (for eval)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	m = finalModel.(model)

	// User quit without selecting
	if m.selected == -1 {
		return nil
	}

	selectedProfile := m.profiles[m.selected]

	// Generate shell commands to stdout
	if err := generateShellCommands(selectedProfile); err != nil {
		return err
	}

	// Save current profile state
	if err := profile.SetCurrentProfile(selectedProfile.Name); err != nil {
		// Non-fatal, just log to stderr
		fmt.Fprintf(os.Stderr, "Warning: failed to save current profile state: %v\n", err)
	}

	return nil
}

func generateShellCommands(p profile.Profile) error {
	var commands []string

	// Always clear Vertex vars to prevent leakage
	commands = append(commands,
		"unset CLAUDE_CODE_USE_VERTEX",
		"unset CLOUD_ML_REGION",
		"unset ANTHROPIC_VERTEX_PROJECT_ID",
	)

	if p.IsVertex {
		// Vertex AI auth: export environment variables
		envVars, err := profile.LoadEnvVars(p.Path)
		if err != nil {
			return fmt.Errorf("failed to load env vars: %w", err)
		}

		if envVars.UseVertex {
			commands = append(commands, "export CLAUDE_CODE_USE_VERTEX=1")
		}
		if envVars.Region != "" {
			commands = append(commands, fmt.Sprintf("export CLOUD_ML_REGION=%s", envVars.Region))
		}
		if envVars.ProjectID != "" {
			commands = append(commands, fmt.Sprintf("export ANTHROPIC_VERTEX_PROJECT_ID=%s", envVars.ProjectID))
		}
	} else {
		// Standard auth: restore token from keyring
		token, err := keyring.GetProfileToken(p.Name)
		if err != nil {
			return err
		}

		if err := keyring.SetActiveToken(token); err != nil {
			return fmt.Errorf("failed to set active token: %w", err)
		}

		// Add a comment to stderr so user knows token was restored
		fmt.Fprintf(os.Stderr, "Restored session token for profile '%s'\n", p.Name)
	}

	// Merge settings.json
	if err := config.MergeSettings(p.Path); err != nil {
		return fmt.Errorf("failed to merge settings: %w", err)
	}

	// Output shell commands
	for _, cmd := range commands {
		fmt.Println(cmd)
	}

	// Success message to stderr
	fmt.Fprintf(os.Stderr, "Switched to Claude profile: %s\n", p.Name)

	return nil
}
