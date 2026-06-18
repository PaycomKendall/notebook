package todo

import (
	"errors"
	"testing"
)

func TestCreateListValidatesAndRejectsDup(t *testing.T) {
	svc := NewService(newFakeRepo())
	if err := svc.CreateList("Bad Name"); !errors.Is(err, ErrInvalidName) {
		t.Errorf("err = %v, want ErrInvalidName", err)
	}
	if err := svc.CreateList("work"); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateList("work"); err != ErrListExists {
		t.Errorf("dup err = %v, want ErrListExists", err)
	}
}

func TestRenameAndDeleteList(t *testing.T) {
	svc := NewService(newFakeRepo())
	_ = svc.CreateList("old")
	if err := svc.RenameList("old", "new"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetList("new"); err != nil {
		t.Errorf("renamed list missing: %v", err)
	}
	if err := svc.DeleteList("new"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetList("new"); err != ErrListNotFound {
		t.Error("deleted list should be gone")
	}
}
