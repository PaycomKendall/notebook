package todo

import "testing"

func TestAddAssignsMonotonicIDs(t *testing.T) {
	l := &List{Name: "inbox"}

	a, err := l.Add("first")
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if a.ID != 1 {
		t.Errorf("first task ID = %d, want 1", a.ID)
	}

	b, err := l.Add("second")
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if b.ID != 2 {
		t.Errorf("second task ID = %d, want 2", b.ID)
	}
	if l.NextID != 3 {
		t.Errorf("NextID = %d, want 3", l.NextID)
	}
	if len(l.Tasks) != 2 {
		t.Errorf("len(Tasks) = %d, want 2", len(l.Tasks))
	}
	if a.Created.IsZero() || a.Updated.IsZero() {
		t.Error("Created/Updated timestamps should be set")
	}
}

func TestAddRejectsEmptyTitle(t *testing.T) {
	l := &List{Name: "inbox"}
	if _, err := l.Add("   "); err != ErrEmptyTitle {
		t.Errorf("Add(blank) err = %v, want ErrEmptyTitle", err)
	}
	if len(l.Tasks) != 0 {
		t.Errorf("blank title should not append; len = %d", len(l.Tasks))
	}
}

func TestNormalizeListName(t *testing.T) {
	got, err := NormalizeListName("  Work ")
	if err != nil || got != "work" {
		t.Errorf("NormalizeListName(' Work ') = %q,%v; want \"work\",nil", got, err)
	}
	if _, err := NormalizeListName("bad name"); err != ErrInvalidName {
		t.Errorf("space name err = %v, want ErrInvalidName", err)
	}
	if _, err := NormalizeListName(""); err != ErrInvalidName {
		t.Errorf("empty err = %v, want ErrInvalidName", err)
	}
}
