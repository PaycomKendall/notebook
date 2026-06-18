package jsonstore

import (
	"testing"

	"github.com/kendallowen/notebook/internal/todo"
)

func TestNamesSortedAndCreate(t *testing.T) {
	s, _ := New(t.TempDir())
	if _, err := s.Create("work"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("home"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("work"); err != todo.ErrListExists {
		t.Errorf("dup Create = %v, want ErrListExists", err)
	}
	names, err := s.Names()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "home" || names[1] != "work" {
		t.Errorf("Names = %v, want [home work]", names)
	}
}

func TestDeleteAndRename(t *testing.T) {
	s, _ := New(t.TempDir())
	_, _ = s.Create("old")
	if err := s.Rename("old", "new"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load("new"); err != nil {
		t.Errorf("renamed list missing: %v", err)
	}
	if _, err := s.Load("old"); err != todo.ErrListNotFound {
		t.Error("old name should be gone")
	}
	if err := s.Delete("new"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("new"); err != todo.ErrListNotFound {
		t.Errorf("Delete(missing) = %v, want ErrListNotFound", err)
	}
}

// Compile-time check that Store implements the port.
var _ todo.ListRepository = (*Store)(nil)
