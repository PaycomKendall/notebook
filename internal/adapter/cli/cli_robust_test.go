package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

// A list file whose on-disk stem is not a valid canonical name (e.g. a
// hand-placed file in an NB_DIR git repo) must be skipped by `lists` and
// `ls --all`, not abort the whole command.
func TestListsAndLsAllSkipUnloadableFiles(t *testing.T) {
	dir := t.TempDir()
	store, err := jsonstore.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	svc := todo.NewService(store)
	run := func(args ...string) (string, error) {
		cmd := NewRootCmd(svc, func(string, string) error { return nil })
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs(args)
		e := cmd.Execute()
		return buf.String(), e
	}
	if _, err := run("add", "real task", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	// Stem "weird name" has a space -> GetList normalizes-fails -> ErrInvalidName.
	if err := os.WriteFile(filepath.Join(dir, "weird name.json"),
		[]byte(`{"name":"weird name","next_id":1,"tasks":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := run("lists")
	if err != nil {
		t.Fatalf("lists aborted on an unloadable file: %v", err)
	}
	if !strings.Contains(out, "work") {
		t.Errorf("lists should still show the valid list; got %q", out)
	}

	out, err = run("ls", "-a")
	if err != nil {
		t.Fatalf("ls --all aborted on an unloadable file: %v", err)
	}
	if !strings.Contains(out, "real task") {
		t.Errorf("ls --all should still show the valid task; got %q", out)
	}
}
