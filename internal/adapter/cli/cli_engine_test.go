package cli

import (
	"bytes"
	"testing"

	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

func runEngine(t *testing.T, env string, args ...string) (engine string, err error) {
	t.Helper()
	t.Setenv("NB_TUI", env)
	store, e := jsonstore.New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	svc := todo.NewService(store)
	got := ""
	cmd := NewRootCmd(svc, func(eng, _ string) error { got = eng; return nil })
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	return got, cmd.Execute()
}

func TestEngineResolution(t *testing.T) {
	if e, err := runEngine(t, "", "tui"); err != nil || e != "tview" {
		t.Errorf("default = %q (err %v), want tview", e, err)
	}
	if e, err := runEngine(t, "", "tui", "--engine", "bubble"); err != nil || e != "bubble" {
		t.Errorf("flag = %q (err %v), want bubble", e, err)
	}
	if e, err := runEngine(t, "bubble", "tui"); err != nil || e != "bubble" {
		t.Errorf("env = %q (err %v), want bubble", e, err)
	}
	if e, err := runEngine(t, "bubble", "tui", "-e", "tview"); err != nil || e != "tview" {
		t.Errorf("flag-overrides-env = %q (err %v), want tview", e, err)
	}
	if _, err := runEngine(t, "", "tui", "--engine", "nope"); err == nil {
		t.Error("invalid engine should error")
	}
}
