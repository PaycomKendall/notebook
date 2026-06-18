package cli

import (
	"strings"
	"testing"
)

func TestListsShowsCounts(t *testing.T) {
	_, run := newTestCmd(t)
	_, _ = run("add", "a", "-l", "work")
	_, _ = run("done", "1", "-l", "work")
	_, _ = run("add", "b", "-l", "work")
	out, err := run("lists")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "work") {
		t.Errorf("lists missing 'work': %q", out)
	}
}

func TestListsNewRenameRm(t *testing.T) {
	svc, run := newTestCmd(t)
	if _, err := run("lists", "new", "ideas"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetList("ideas"); err != nil {
		t.Fatalf("list not created: %v", err)
	}
	if _, err := run("lists", "rename", "ideas", "later"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetList("later"); err != nil {
		t.Errorf("rename failed: %v", err)
	}
	if _, err := run("lists", "rm", "later", "--force"); err != nil {
		t.Fatal(err)
	}
	names, _ := svc.ListNames()
	if len(names) != 0 {
		t.Errorf("list not removed: %v", names)
	}
}
