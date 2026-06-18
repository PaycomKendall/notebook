package todo

import "testing"

func svcWithTask(t *testing.T) *Service {
	t.Helper()
	svc := NewService(newFakeRepo())
	if _, err := svc.AddTask("work", "ship it", nil, ""); err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestToggleTaskPersists(t *testing.T) {
	svc := svcWithTask(t)
	if err := svc.ToggleTask("work", 1); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if !l.Tasks[0].Done {
		t.Error("toggle did not persist")
	}
}

func TestEditTaskPartialUpdate(t *testing.T) {
	svc := svcWithTask(t)
	newNotes := "with details"
	if err := svc.EditTask("work", 1, nil, &newNotes); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if l.Tasks[0].Title != "ship it" {
		t.Error("title should be unchanged when nil")
	}
	if l.Tasks[0].Notes != "with details" {
		t.Error("notes not updated")
	}
}

func TestMutationOnMissingList(t *testing.T) {
	svc := NewService(newFakeRepo())
	if err := svc.ToggleTask("nope", 1); err != ErrListNotFound {
		t.Errorf("err = %v, want ErrListNotFound", err)
	}
}

func TestTaskTagThroughService(t *testing.T) {
	svc := svcWithTask(t)
	if err := svc.AddTaskTag("work", 1, "Urgent"); err != nil {
		t.Fatal(err)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks[0].Tags) != 1 || l.Tasks[0].Tags[0] != "urgent" {
		t.Errorf("tags = %v", l.Tasks[0].Tags)
	}
	if err := svc.RemoveTaskTag("work", 1, "urgent"); err != nil {
		t.Fatal(err)
	}
	l, _ = svc.GetList("work")
	if len(l.Tasks[0].Tags) != 0 {
		t.Errorf("tags after remove = %v", l.Tasks[0].Tags)
	}
}
