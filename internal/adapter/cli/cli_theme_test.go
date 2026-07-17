package cli

import (
	"bytes"
	"testing"

	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

func runTheme(t *testing.T, env string, args ...string) (theme string, err error) {
	t.Helper()
	t.Setenv("NB_THEME", env)
	store, e := jsonstore.New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	svc := todo.NewService(store)
	got := ""
	cmd := NewRootCmd(svc, func(th string) error { got = th; return nil })
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	return got, cmd.Execute()
}

func TestThemeResolution(t *testing.T) {
	if th, err := runTheme(t, "", "tui"); err != nil || th != "default" {
		t.Errorf("default = %q (err %v), want default", th, err)
	}
	if th, err := runTheme(t, "", "tui", "--theme", "nord"); err != nil || th != "nord" {
		t.Errorf("flag = %q (err %v), want nord", th, err)
	}
	if th, err := runTheme(t, "dracula", "tui"); err != nil || th != "dracula" {
		t.Errorf("env = %q (err %v), want dracula", th, err)
	}
	if th, err := runTheme(t, "dracula", "tui", "--theme", "mono"); err != nil || th != "mono" {
		t.Errorf("flag-overrides-env = %q (err %v), want mono", th, err)
	}
}
