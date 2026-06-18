package cli

import (
	"os"

	"github.com/kendallowen/notebook/internal/todo"
	"github.com/spf13/cobra"
)

// resolveList applies the default-list rule: flag, then $NB_LIST, then "inbox".
func resolveList(flag string) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("NB_LIST"); env != "" {
		return env
	}
	return "inbox"
}

// NewRootCmd builds the command tree. launchTUI runs the interactive UI.
func NewRootCmd(svc *todo.Service, launchTUI func() error) *cobra.Command {
	root := &cobra.Command{
		Use:           "nb",
		Short:         "notebook — a CLI + TUI task tracker",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return launchTUI()
		},
	}
	tui := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		RunE:  func(cmd *cobra.Command, args []string) error { return launchTUI() },
	}
	root.AddCommand(tui)
	root.AddCommand(newAddCmd(svc))
	return root
}
