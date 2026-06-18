package cli

import (
	"fmt"

	"github.com/kendallowen/notebook/internal/todo"
	"github.com/spf13/cobra"
)

func newEditCmd(svc *todo.Service) *cobra.Command {
	var list, title, note string
	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a task's title and/or note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			var tp, np *string
			if cmd.Flags().Changed("title") {
				tp = &title
			}
			if cmd.Flags().Changed("note") {
				np = &note
			}
			if err := svc.EditTask(resolveList(list), ids[0], tp, np); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "edited #%d\n", ids[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $NB_LIST)")
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVarP(&note, "note", "n", "", "new note")
	return cmd
}
