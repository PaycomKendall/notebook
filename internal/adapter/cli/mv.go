package cli

import (
	"fmt"
	"strconv"

	"github.com/kendallowen/notebook/internal/todo"
	"github.com/spf13/cobra"
)

func newMvCmd(svc *todo.Service) *cobra.Command {
	var list string
	cmd := &cobra.Command{
		Use:     "mv <id> <dest>",
		Aliases: []string{"move"},
		Short:   "Move a task to another list",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q", args[0])
			}
			dest := args[1]
			moved, err := svc.MoveTask(resolveList(list), id, dest)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "moved %q to %s as #%d\n", moved.Title, dest, moved.ID)
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "source list (default: inbox or $NB_LIST)")
	return cmd
}
