package todo

import (
	"errors"
	"testing"
)

// fakeRepo is an in-memory ListRepository for service tests.
type fakeRepo struct {
	lists map[string]*List
}

func newFakeRepo() *fakeRepo { return &fakeRepo{lists: map[string]*List{}} }

func clone(l *List) *List {
	cp := *l
	cp.Tasks = append([]Task(nil), l.Tasks...)
	return &cp
}

func (f *fakeRepo) Names() ([]string, error) {
	out := make([]string, 0, len(f.lists))
	for name := range f.lists {
		out = append(out, name)
	}
	return out, nil
}
func (f *fakeRepo) Load(name string) (*List, error) {
	l, ok := f.lists[name]
	if !ok {
		return nil, ErrListNotFound
	}
	return clone(l), nil
}
func (f *fakeRepo) Save(l *List) error { f.lists[l.Name] = clone(l); return nil }
func (f *fakeRepo) Create(name string) (*List, error) {
	if _, ok := f.lists[name]; ok {
		return nil, ErrListExists
	}
	l := &List{Name: name, NextID: 1}
	f.lists[name] = clone(l)
	return l, nil
}
func (f *fakeRepo) Delete(name string) error {
	if _, ok := f.lists[name]; !ok {
		return ErrListNotFound
	}
	delete(f.lists, name)
	return nil
}
func (f *fakeRepo) Rename(oldName, newName string) error {
	l, ok := f.lists[oldName]
	if !ok {
		return ErrListNotFound
	}
	if _, ok := f.lists[newName]; ok {
		return ErrListExists
	}
	l.Name = newName
	f.lists[newName] = l
	delete(f.lists, oldName)
	return nil
}

func TestAddTaskAutoCreatesList(t *testing.T) {
	svc := NewService(newFakeRepo())
	task, err := svc.AddTask("groceries", "milk", []string{"Store"}, "2%")
	if err != nil {
		t.Fatalf("AddTask error: %v", err)
	}
	if task.ID != 1 || task.Title != "milk" {
		t.Errorf("task = %+v", task)
	}
	if len(task.Tags) != 1 || task.Tags[0] != "store" {
		t.Errorf("tags = %v, want [store]", task.Tags)
	}
	if task.Notes != "2%" {
		t.Errorf("notes = %q", task.Notes)
	}
	l, err := svc.GetList("groceries")
	if err != nil {
		t.Fatalf("list should have been created: %v", err)
	}
	if len(l.Tasks) != 1 {
		t.Errorf("persisted task count = %d, want 1", len(l.Tasks))
	}
}

func TestAddTaskRejectsBadListName(t *testing.T) {
	svc := NewService(newFakeRepo())
	if _, err := svc.AddTask("Bad Name", "x", nil, ""); !errors.Is(err, ErrInvalidName) {
		t.Errorf("err = %v, want ErrInvalidName", err)
	}
}

func TestAddTaskNormalizesListName(t *testing.T) {
	svc := NewService(newFakeRepo())
	if _, err := svc.AddTask("Work", "ship it", nil, ""); err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	l, err := svc.GetList("work") // mixed-case must resolve to canonical "work"
	if err != nil {
		t.Fatalf("GetList(work): %v", err)
	}
	if l.Name != "work" {
		t.Errorf("stored List.Name = %q, want \"work\"", l.Name)
	}
	if len(l.Tasks) != 1 {
		t.Fatalf("want 1 task in canonical list, got %d", len(l.Tasks))
	}
	// A different-case form must hit the SAME list, not create a new one.
	if _, err := svc.AddTask("WORK", "second", nil, ""); err != nil {
		t.Fatal(err)
	}
	l, _ = svc.GetList("work")
	if len(l.Tasks) != 2 {
		t.Errorf("WORK/Work/work must be one list; got %d tasks", len(l.Tasks))
	}
	names, _ := svc.ListNames()
	if len(names) != 1 {
		t.Errorf("expected exactly one underlying list, got %v", names)
	}
}
