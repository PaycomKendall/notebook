package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kendallowen/notebook/internal/todo"
	"github.com/spf13/cobra"
)

// parseIDs converts string args to ints.
func parseIDs(args []string) ([]int, error) {
	ids := make([]int, 0, len(args))
	for _, a := range args {
		n, err := strconv.Atoi(a)
		if err != nil {
			return nil, fmt.Errorf("invalid id %q", a)
		}
		ids = append(ids, n)
	}
	return ids, nil
}

func formatTask(t todo.Task) string {
	box := "[ ]"
	if t.Done {
		box = "[x]"
	}
	line := fmt.Sprintf("%s #%d %s", box, t.ID, t.Title)
	if len(t.Tags) > 0 {
		line += "  #" + strings.Join(t.Tags, " #")
	}
	return line
}

func newLsCmd(svc *todo.Service) *cobra.Command {
	var list, tag string
	var all, doneOnly, openOnly bool
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			var names []string
			if all {
				n, err := svc.ListNames()
				if err != nil {
					return err
				}
				names = n
			} else {
				names = []string{resolveList(list)}
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
				fmt.Fprintf(out, "%s:\n", name)
				for _, task := range l.Tasks {
					if tag != "" && !hasTag(task, tag) {
						continue
					}
					if doneOnly && !task.Done {
						continue
					}
					if openOnly && task.Done {
						continue
					}
					fmt.Fprintf(out, "  %s\n", formatTask(task))
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&list, "list", "l", "", "list name (default: inbox or $NB_LIST)")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "show all lists")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "filter by tag")
	cmd.Flags().BoolVar(&doneOnly, "done", false, "show only done tasks")
	cmd.Flags().BoolVar(&openOnly, "open", false, "show only open tasks")
	return cmd
}

func hasTag(t todo.Task, tag string) bool {
	tag = strings.ToLower(strings.TrimSpace(tag))
	for _, x := range t.Tags {
		if x == tag {
			return true
		}
	}
	return false
}
