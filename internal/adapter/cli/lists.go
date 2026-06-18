package cli

import (
	"errors"
	"fmt"

	"github.com/kendallowen/notebook/internal/todo"
	"github.com/spf13/cobra"
)

func newListsCmd(svc *todo.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lists",
		Short: "Manage lists",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := svc.ListNames()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, name := range names {
				l, err := svc.GetList(name)
				if errors.Is(err, todo.ErrListNotFound) || errors.Is(err, todo.ErrInvalidName) {
					continue
				}
				if err != nil {
					return err
				}
				open, done := 0, 0
				for _, task := range l.Tasks {
					if task.Done {
						done++
					} else {
						open++
					}
				}
				fmt.Fprintf(out, "%s (%d open, %d done)\n", name, open, done)
			}
			return nil
		},
	}

	newCmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.CreateList(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created list %q\n", args[0])
			return nil
		},
	}

	var force bool
	rmCmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Delete a list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return fmt.Errorf("refusing to delete %q without --force", args[0])
			}
			if err := svc.DeleteList(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted list %q\n", args[0])
			return nil
		},
	}
	rmCmd.Flags().BoolVar(&force, "force", false, "confirm deletion")

	renameCmd := &cobra.Command{
		Use:   "rename <old> <new>",
		Short: "Rename a list",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.RenameList(args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "renamed %q to %q\n", args[0], args[1])
			return nil
		},
	}

	cmd.AddCommand(newCmd, rmCmd, renameCmd)
	return cmd
}
