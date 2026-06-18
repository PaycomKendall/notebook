package cli

import (
	"fmt"

	"github.com/kendallowen/notebook/internal/todo"
	"github.com/spf13/cobra"
)

func newDoneCmd(svc *todo.Service) *cobra.Command   { return doneLikeCmd(svc, "done", true) }
func newUndoneCmd(svc *todo.Service) *cobra.Command { return doneLikeCmd(svc, "undone", false) }

func doneLikeCmd(svc *todo.Service, use string, done bool) *cobra.Command {
	var list string
	cmd := &cobra.Command{
		Use:   use + " <id>...",
		Short: "Mark task(s) " + use,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			name := resolveList(list)
			for _, id := range ids {
				if err := svc.SetTaskDone(name, id, done); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s #%d\n", use, id)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $NB_LIST)")
	return cmd
}
