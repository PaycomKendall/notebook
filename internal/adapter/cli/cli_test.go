package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

// newTestCmd builds a root command backed by a temp-dir store.
func newTestCmd(t *testing.T) (*todo.Service, func(args ...string) (string, error)) {
	t.Helper()
	store, err := jsonstore.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := todo.NewService(store)
	run := func(args ...string) (string, error) {
		cmd := NewRootCmd(svc, func(string) error { return nil })
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs(args)
		err := cmd.Execute()
		return buf.String(), err
	}
	return svc, run
}

func TestAddCommand(t *testing.T) {
	svc, run := newTestCmd(t)
	out, err := run("add", "buy milk", "-l", "groceries", "-t", "store")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	if !strings.Contains(out, "#1") || !strings.Contains(out, "buy milk") {
		t.Errorf("unexpected output: %q", out)
	}
	l, err := svc.GetList("groceries")
	if err != nil {
		t.Fatalf("list not created: %v", err)
	}
	if len(l.Tasks) != 1 || l.Tasks[0].Tags[0] != "store" {
		t.Errorf("task not persisted correctly: %+v", l.Tasks)
	}
}
