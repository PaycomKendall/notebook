package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

// runRoot builds a fresh root command over a temp-dir store and a launchTUI
// spy, runs it with args, and returns the output plus whether the TUI launched.
func runRoot(t *testing.T, args ...string) (out string, tuiLaunched bool, err error) {
	t.Helper()
	store, e := jsonstore.New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	svc := todo.NewService(store)
	launched := false
	cmd := NewRootCmd(svc, func(string, string) error { launched = true; return nil })
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	e = cmd.Execute()
	return buf.String(), launched, e
}

func TestBareRootPrintsHelpWithBannerAndDoesNotLaunchTUI(t *testing.T) {
	out, launched, err := runRoot(t)
	if err != nil {
		t.Fatalf("bare nb returned error: %v", err)
	}
	if launched {
		t.Error("bare `nb` must NOT launch the TUI")
	}
	// A distinctive fragment of the Modular "Notebook" banner.
	if !strings.Contains(out, "|_|  |__|") {
		t.Errorf("help output should contain the Notebook banner; got:\n%s", out)
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("bare `nb` should print help/usage; got:\n%s", out)
	}
}

func TestTuiSubcommandLaunchesTUI(t *testing.T) {
	_, launched, err := runRoot(t, "tui")
	if err != nil {
		t.Fatalf("nb tui returned error: %v", err)
	}
	if !launched {
		t.Error("`nb tui` must launch the TUI")
	}
}

func TestHelpFlagShowsBanner(t *testing.T) {
	out, launched, err := runRoot(t, "--help")
	if err != nil {
		t.Fatalf("nb --help returned error: %v", err)
	}
	if launched {
		t.Error("`nb --help` must NOT launch the TUI")
	}
	if !strings.Contains(out, "|_|  |__|") {
		t.Errorf("`nb --help` should contain the Notebook banner; got:\n%s", out)
	}
}
