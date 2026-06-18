package todo

import "testing"

func TestAppendPreservesFieldsWithNewID(t *testing.T) {
	src := &List{Name: "src"}
	orig, _ := src.Add("buy milk")
	_ = src.SetDone(orig.ID, true)
	_ = src.AddTag(orig.ID, "store")
	_ = src.SetNotes(orig.ID, "2%")
	moved := src.Tasks[0]

	dest := &List{Name: "dest", NextID: 5}
	got := dest.Append(moved)
	if got.ID != 5 || dest.NextID != 6 {
		t.Errorf("id/nextid: got.ID=%d dest.NextID=%d, want 5/6", got.ID, dest.NextID)
	}
	if got.Title != "buy milk" || !got.Done || got.Notes != "2%" ||
		len(got.Tags) != 1 || got.Tags[0] != "store" {
		t.Errorf("fields not preserved: %+v", got)
	}
	if !got.Created.Equal(moved.Created) {
		t.Error("Created should be preserved")
	}
}

func TestAppendClonesTags(t *testing.T) {
	src := &List{Name: "src"}
	a, _ := src.Add("a")
	_ = src.AddTag(a.ID, "x")
	moved := src.Tasks[0]
	dest := &List{Name: "dest"}
	got := dest.Append(moved)
	got.Tags[0] = "mutated"
	if moved.Tags[0] != "x" {
		t.Error("Append must clone Tags so the source isn't aliased")
	}
}
