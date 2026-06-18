package cli

import (
	"fmt"
	"strings"

	"github.com/kendallowen/notebook/internal/todo"
	"github.com/spf13/cobra"
)

func newAddCmd(svc *todo.Service) *cobra.Command {
	var list, note string
	var tags []string
	cmd := &cobra.Command{
		Use:   "add <title>...",
		Short: "Add a task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := resolveList(list)
			title := strings.Join(args, " ")
			task, err := svc.AddTask(name, title, tags, note)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added [%s] #%d: %s\n", name, task.ID, task.Title)
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $NB_LIST)")
	cmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "tag (repeatable or comma-separated)")
	cmd.Flags().StringVarP(&note, "note", "n", "", "note text")
	return cmd
}
