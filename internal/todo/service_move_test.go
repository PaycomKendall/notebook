package todo

import "testing"

func TestMoveTaskBetweenLists(t *testing.T) {
	svc := NewService(newFakeRepo())
	added, _ := svc.AddTask("work", "ship it", []string{"urgent"}, "soon")
	moved, err := svc.MoveTask("work", added.ID, "archive")
	if err != nil {
		t.Fatal(err)
	}
	work, _ := svc.GetList("work")
	if len(work.Tasks) != 0 {
		t.Errorf("task still in source: %+v", work.Tasks)
	}
	arch, err := svc.GetList("archive")
	if err != nil {
		t.Fatalf("dest not auto-created: %v", err)
	}
	if len(arch.Tasks) != 1 {
		t.Fatalf("dest task count = %d", len(arch.Tasks))
	}
	got := arch.Tasks[0]
	if got.Title != "ship it" || got.Notes != "soon" || len(got.Tags) != 1 {
		t.Errorf("fields not preserved: %+v", got)
	}
	if got.ID != moved.ID {
		t.Errorf("returned id %d != stored id %d", moved.ID, got.ID)
	}
}

func TestMoveTaskSameListIsNoOp(t *testing.T) {
	svc := NewService(newFakeRepo())
	added, _ := svc.AddTask("work", "stay", nil, "")
	moved, err := svc.MoveTask("work", added.ID, "work")
	if err != nil {
		t.Fatal(err)
	}
	if moved.ID != added.ID {
		t.Errorf("no-op should keep id %d, got %d", added.ID, moved.ID)
	}
	work, _ := svc.GetList("work")
	if len(work.Tasks) != 1 {
		t.Errorf("no-op changed task count: %d", len(work.Tasks))
	}
}

func TestMoveTaskErrors(t *testing.T) {
	svc := NewService(newFakeRepo())
	_, _ = svc.AddTask("work", "x", nil, "")
	if _, err := svc.MoveTask("nope", 1, "archive"); err != ErrListNotFound {
		t.Errorf("missing src = %v, want ErrListNotFound", err)
	}
	if _, err := svc.MoveTask("work", 99, "archive"); err != ErrTaskNotFound {
		t.Errorf("missing task = %v, want ErrTaskNotFound", err)
	}
}
