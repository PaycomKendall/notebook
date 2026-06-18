package cli

import (
	"strings"
	"testing"
)

func TestMvMovesTask(t *testing.T) {
	svc, run := newTestCmd(t)
	_, _ = run("add", "buy milk", "-l", "work")
	out, err := run("mv", "1", "groceries", "-l", "work")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "groceries") || !strings.Contains(out, "buy milk") {
		t.Errorf("unexpected output: %q", out)
	}
	work, _ := svc.GetList("work")
	if len(work.Tasks) != 0 {
		t.Errorf("task still in source: %+v", work.Tasks)
	}
	groc, err := svc.GetList("groceries")
	if err != nil || len(groc.Tasks) != 1 {
		t.Fatalf("dest not created/populated: err=%v list=%+v", err, groc)
	}
}

func TestMvAliasAndBadID(t *testing.T) {
	_, run := newTestCmd(t)
	_, _ = run("add", "x", "-l", "work")
	if _, err := run("move", "1", "done", "-l", "work"); err != nil {
		t.Fatalf("alias move failed: %v", err)
	}
	if _, err := run("mv", "notanint", "done", "-l", "work"); err == nil {
		t.Error("non-numeric id should error")
	}
}
