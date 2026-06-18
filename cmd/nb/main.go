package main

import (
	"fmt"
	"os"

	"github.com/kendallowen/notebook/internal/adapter/cli"
	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

func main() {
	dir := os.Getenv("NB_DIR")
	store, err := jsonstore.New(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	svc := todo.NewService(store)

	// Placeholder until the TUI adapter lands in Task 16.
	launchTUI := func() error {
		fmt.Fprintln(os.Stderr, "TUI not yet implemented — use subcommands (nb --help)")
		return nil
	}

	root := cli.NewRootCmd(svc, launchTUI)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
