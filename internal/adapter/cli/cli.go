package cli

import (
	"fmt"
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

// resolveEngine applies the engine rule: flag, then $NB_TUI, then "tview".
func resolveEngine(flag string) (string, error) {
	e := flag
	if e == "" {
		e = os.Getenv("NB_TUI")
	}
	if e == "" {
		e = "tview"
	}
	switch e {
	case "tview", "bubble":
		return e, nil
	default:
		return "", fmt.Errorf("invalid engine %q (want \"tview\" or \"bubble\")", e)
	}
}

// NewRootCmd builds the command tree. launchTUI runs the interactive UI for
// the chosen engine; bare `nb` prints help.
func NewRootCmd(svc *todo.Service, launchTUI func(engine, theme string) error) *cobra.Command {
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
	var engine, theme string
	tui := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := resolveEngine(engine)
			if err != nil {
				return err
			}
			th := theme
			if th == "" {
				th = os.Getenv("NB_THEME")
			}
			if th == "" {
				th = "default"
			}
			return launchTUI(eng, th)
		},
	}
	tui.Flags().StringVarP(&engine, "engine", "e", "", `TUI engine: "tview" (default) or "bubble"; or $NB_TUI`)
	tui.Flags().StringVar(&theme, "theme", "", `bubble theme: default, nord, dracula, gruvbox, mono; or $NB_THEME`)
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
