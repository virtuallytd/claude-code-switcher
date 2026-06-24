package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "ccs <profile> [path] [-- claude-args...]",
		Short:   "Claude Code Switcher — run Claude Code with isolated Podman profiles",
		Version: version,
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
	root.AddCommand(initCmd())
	root.AddCommand(runCmd())

	knownCmds := map[string]bool{"build": true, "stop": true, "status": true, "profiles": true, "init": true, "run": true, "help": true, "completion": true}
	if len(os.Args) > 1 && !knownCmds[os.Args[1]] && !strings.HasPrefix(os.Args[1], "-") {
		os.Args = append([]string{os.Args[0], "run"}, os.Args[1:]...)
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
