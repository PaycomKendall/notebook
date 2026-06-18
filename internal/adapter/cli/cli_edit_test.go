package cli

import "testing"

func TestEditUpdatesTitleAndNote(t *testing.T) {
	svc, run := newTestCmd(t)
	_, _ = run("add", "old title", "-l", "work")
	if _, err := run("edit", "1", "--title", "new title", "-n", "a note", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if l.Tasks[0].Title != "new title" || l.Tasks[0].Notes != "a note" {
		t.Errorf("edit failed: %+v", l.Tasks[0])
	}
}

func TestTagAddAndRemove(t *testing.T) {
	svc, run := newTestCmd(t)
	_, _ = run("add", "task", "-l", "work")
	if _, err := run("tag", "1", "--add", "urgent", "--add", "home", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks[0].Tags) != 2 {
		t.Fatalf("tags = %v", l.Tasks[0].Tags)
	}
	if _, err := run("tag", "1", "--rm", "home", "-l", "work"); err != nil {
		t.Fatal(err)
	}
	l, _ = svc.GetList("work")
	if len(l.Tasks[0].Tags) != 1 || l.Tasks[0].Tags[0] != "urgent" {
		t.Errorf("tags after rm = %v", l.Tasks[0].Tags)
	}
}
