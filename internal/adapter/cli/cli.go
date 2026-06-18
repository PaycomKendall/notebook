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

// banner is the Modular-font "Notebook" wordmark shown atop the help output.
const banner = `
 __    _  _______  _______  _______  _______  _______  _______  ___   _
|  |  | ||       ||       ||       ||  _    ||       ||       ||   | | |
|   |_| ||   _   ||_     _||    ___|| |_|   ||   _   ||   _   ||   |_| |
|       ||  | |  |  |   |  |   |___ |       ||  | |  ||  | |  ||      _|
|  _    ||  |_|  |  |   |  |    ___||  _   | |  |_|  ||  |_|  ||     |_
| | |   ||       |  |   |  |   |___ | |_|   ||       ||       ||    _  |
|_|  |__||_______|  |___|  |_______||_______||_______||_______||___| |_|`

// NewRootCmd builds the command tree. launchTUI runs the interactive UI,
// reached only via `nb tui`; bare `nb` prints help.
func NewRootCmd(svc *todo.Service, launchTUI func() error) *cobra.Command {
	root := &cobra.Command{
		Use:           "nb",
		Short:         "notebook — a CLI + TUI task tracker",
		Long:          banner + "\n\nnotebook — a CLI + TUI task tracker.\nRun `nb tui` for the interactive interface.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	tui := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		RunE:  func(cmd *cobra.Command, args []string) error { return launchTUI() },
	}
	root.AddCommand(tui)
	root.AddCommand(newAddCmd(svc))
	root.AddCommand(newLsCmd(svc))
	root.AddCommand(newDoneCmd(svc))
	root.AddCommand(newUndoneCmd(svc))
	root.AddCommand(newRmCmd(svc))
	root.AddCommand(newEditCmd(svc))
	root.AddCommand(newTagCmd(svc))
	root.AddCommand(newListsCmd(svc))
	return root
}
