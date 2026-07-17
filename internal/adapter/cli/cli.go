package cli

import (
	"fmt"
	"os"
	"strings"

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

// wordmark is the Modular-font "Notebook" lettering shown atop the help output.
var wordmark = []string{
	` __    _  _______  _______  _______  _______  _______  _______  ___   _`,
	`|  |  | ||       ||       ||       ||  _    ||       ||       ||   | | |`,
	`|   |_| ||   _   ||_     _||    ___|| |_|   ||   _   ||   _   ||   |_| |`,
	`|       ||  | |  |  |   |  |   |___ |       ||  | |  ||  | |  ||      _|`,
	`|  _    ||  |_|  |  |   |  |    ___||  _   | |  |_|  ||  |_|  ||     |_`,
	`| | |   ||       |  |   |  |   |___ | |_|   ||       ||       ||    _  |`,
	`|_|  |__||_______|  |___|  |_______||_______||_______||_______||___| |_|`,
}

// notebookIcon is a small spiral notepad drawn to the right of the wordmark.
var notebookIcon = []string{
	` .---------.`,
	` |o o o o o|`,
	` |=========|`,
	` | ------- |`,
	` | ------- |`,
	` | ------- |`,
	` '---------'`,
}

// banner joins the wordmark and the notebook icon side by side, prefixed with a
// blank line (matching the original leading newline in the help output).
var banner = buildBanner()

func buildBanner() string {
	w := 0
	for _, line := range wordmark {
		if len(line) > w { // ASCII-only, so byte length is the visual width
			w = len(line)
		}
	}
	var b strings.Builder
	b.WriteByte('\n')
	for i, line := range wordmark {
		icon := ""
		if i < len(notebookIcon) {
			icon = notebookIcon[i]
		}
		fmt.Fprintf(&b, "%-*s  %s\n", w, line, icon)
	}
	return strings.TrimRight(b.String(), "\n")
}

// NewRootCmd builds the command tree. launchTUI runs the interactive UI;
// bare `nb` prints help.
func NewRootCmd(svc *todo.Service, launchTUI func(theme string) error) *cobra.Command {
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
	var theme string
	tui := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			th := theme
			if th == "" {
				th = os.Getenv("NB_THEME")
			}
			if th == "" {
				th = "default"
			}
			return launchTUI(th)
		},
	}
	tui.Flags().StringVar(&theme, "theme", "", `theme: default, nord, dracula, gruvbox, mono, notebook, notebook-dark; or $NB_THEME`)
	root.AddCommand(tui)
	root.AddCommand(newAddCmd(svc))
	root.AddCommand(newLsCmd(svc))
	root.AddCommand(newDoneCmd(svc))
	root.AddCommand(newUndoneCmd(svc))
	root.AddCommand(newRmCmd(svc))
	root.AddCommand(newEditCmd(svc))
	root.AddCommand(newTagCmd(svc))
	root.AddCommand(newMvCmd(svc))
	root.AddCommand(newListsCmd(svc))
	return root
}
