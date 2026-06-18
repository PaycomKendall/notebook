package jsonstore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kendallowen/notebook/internal/todo"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	l := &todo.List{Name: "work", NextID: 2,
		Tasks: []todo.Task{{ID: 1, Title: "ship", Tags: []string{"urgent"}, Notes: "go"}}}
	if err := s.Save(l); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load("work")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "work" || got.NextID != 2 || len(got.Tasks) != 1 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	if got.Tasks[0].Title != "ship" || got.Tasks[0].Tags[0] != "urgent" {
		t.Errorf("task mismatch: %+v", got.Tasks[0])
	}
}

func TestLoadMissingIsErrListNotFound(t *testing.T) {
	s, _ := New(t.TempDir())
	if _, err := s.Load("nope"); err != todo.ErrListNotFound {
		t.Errorf("err = %v, want ErrListNotFound", err)
	}
}

func TestSaveIsAtomicNoTempLeft(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	_ = s.Save(&todo.List{Name: "x", NextID: 1})
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestDefaultDirHonorsNBDIR(t *testing.T) {
	t.Setenv("NB_DIR", "/tmp/custom-nb")
	d, err := DefaultDir()
	if err != nil {
		t.Fatal(err)
	}
	if d != "/tmp/custom-nb" {
		t.Errorf("DefaultDir = %q, want /tmp/custom-nb", d)
	}
}

func TestLoadMalformedJSONErrors(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := s.Load("broken")
	if err == nil {
		t.Fatal("expected error loading malformed JSON")
	}
	if err == todo.ErrListNotFound {
		t.Error("malformed JSON should not be reported as not-found")
	}
}
