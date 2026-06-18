package todo

import (
	"reflect"
	"testing"
)

func TestSetTitleAndNotes(t *testing.T) {
	l := seeded(t)
	if err := l.SetTitle(1, "  renamed  "); err != nil {
		t.Fatal(err)
	}
	got, _ := l.Get(1)
	if got.Title != "renamed" {
		t.Errorf("Title = %q, want %q", got.Title, "renamed")
	}
	if err := l.SetTitle(1, "   "); err != ErrEmptyTitle {
		t.Errorf("SetTitle(blank) = %v, want ErrEmptyTitle", err)
	}
	if err := l.SetNotes(1, "buy milk"); err != nil {
		t.Fatal(err)
	}
	got, _ = l.Get(1)
	if got.Notes != "buy milk" {
		t.Errorf("Notes = %q", got.Notes)
	}
}

func TestTagAddNormalizesAndDedupes(t *testing.T) {
	l := seeded(t)
	for _, tag := range []string{"Urgent", " urgent ", "home"} {
		if err := l.AddTag(1, tag); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := l.Get(1)
	if !reflect.DeepEqual(got.Tags, []string{"urgent", "home"}) {
		t.Errorf("Tags = %v, want [urgent home]", got.Tags)
	}
	if err := l.AddTag(1, "  "); err != nil {
		t.Fatalf("blank tag should be a no-op, got %v", err)
	}
	got, _ = l.Get(1)
	if len(got.Tags) != 2 {
		t.Errorf("blank tag changed count: %v", got.Tags)
	}
}

func TestRemoveTag(t *testing.T) {
	l := seeded(t)
	_ = l.AddTag(1, "home")
	_ = l.AddTag(1, "urgent")
	if err := l.RemoveTag(1, "HOME"); err != nil {
		t.Fatal(err)
	}
	got, _ := l.Get(1)
	if !reflect.DeepEqual(got.Tags, []string{"urgent"}) {
		t.Errorf("Tags = %v, want [urgent]", got.Tags)
	}
}
