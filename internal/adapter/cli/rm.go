package cli

import (
	"fmt"

	"github.com/kendallowen/notebook/internal/todo"
	"github.com/spf13/cobra"
)

func newRmCmd(svc *todo.Service) *cobra.Command {
	var list string
	cmd := &cobra.Command{
		Use:   "rm <id>...",
		Short: "Delete task(s)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			name := resolveList(list)
			for _, id := range ids {
				if err := svc.RemoveTask(name, id); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "removed #%d\n", id)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $NB_LIST)")
	return cmd
}
