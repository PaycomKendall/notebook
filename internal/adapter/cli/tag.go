package cli

import (
	"fmt"

	"github.com/kendallowen/notebook/internal/todo"
	"github.com/spf13/cobra"
)

func newTagCmd(svc *todo.Service) *cobra.Command {
	var list string
	var add, rm []string
	cmd := &cobra.Command{
		Use:   "tag <id>",
		Short: "Add or remove tags on a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			name := resolveList(list)
			for _, tag := range add {
				if err := svc.AddTaskTag(name, ids[0], tag); err != nil {
					return err
				}
			}
			for _, tag := range rm {
				if err := svc.RemoveTaskTag(name, ids[0], tag); err != nil {
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated tags on #%d\n", ids[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $NB_LIST)")
	cmd.Flags().StringSliceVar(&add, "add", nil, "tag to add (repeatable)")
	cmd.Flags().StringSliceVar(&rm, "rm", nil, "tag to remove (repeatable)")
	return cmd
}
