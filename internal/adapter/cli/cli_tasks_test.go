package cli

import (
	"strings"
	"testing"
)

func TestLsShowsTasks(t *testing.T) {
	_, run := newTestCmd(t)
	_, _ = run("add", "alpha", "-l", "work")
	_, _ = run("add", "beta", "-l", "work")
	out, err := run("ls", "-l", "work")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("ls missing tasks: %q", out)
	}
	if !strings.Contains(out, "[ ]") {
		t.Errorf("ls missing open checkbox: %q", out)
	}
}

func TestDoneAndUndone(t *testing.T) {
	svc, run := newTestCmd(t)
	_, _ = run("add", "alpha", "-l", "work")
	if _, err := run("done", "1", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if !l.Tasks[0].Done {
		t.Error("done did not mark task")
	}
	if _, err := run("undone", "1", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ = svc.GetList("work")
	if l.Tasks[0].Done {
		t.Error("undone did not clear task")
	}
}

func TestRmDeletesTask(t *testing.T) {
	svc, run := newTestCmd(t)
	_, _ = run("add", "alpha", "-l", "work")
	if _, err := run("rm", "1", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks) != 0 {
		t.Errorf("task not removed: %+v", l.Tasks)
	}
}
