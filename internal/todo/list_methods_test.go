package todo

import "testing"

func seeded(t *testing.T) *List {
	t.Helper()
	l := &List{Name: "inbox"}
	if _, err := l.Add("alpha"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Add("beta"); err != nil {
		t.Fatal(err)
	}
	return l
}

func TestToggleFlipsDone(t *testing.T) {
	l := seeded(t)
	if err := l.Toggle(1); err != nil {
		t.Fatal(err)
	}
	got, _ := l.Get(1)
	if !got.Done {
		t.Error("Toggle should set Done=true")
	}
	if err := l.Toggle(1); err != nil {
		t.Fatal(err)
	}
	got, _ = l.Get(1)
	if got.Done {
		t.Error("second Toggle should set Done=false")
	}
}

func TestSetDoneAndMissingID(t *testing.T) {
	l := seeded(t)
	if err := l.SetDone(2, true); err != nil {
		t.Fatal(err)
	}
	got, _ := l.Get(2)
	if !got.Done {
		t.Error("SetDone(true) failed")
	}
	if err := l.Toggle(99); err != ErrTaskNotFound {
		t.Errorf("Toggle(missing) = %v, want ErrTaskNotFound", err)
	}
	if _, err := l.Get(99); err != ErrTaskNotFound {
		t.Errorf("Get(missing) = %v, want ErrTaskNotFound", err)
	}
}

func TestRemove(t *testing.T) {
	l := seeded(t)
	if err := l.Remove(1); err != nil {
		t.Fatal(err)
	}
	if len(l.Tasks) != 1 {
		t.Fatalf("len after remove = %d, want 1", len(l.Tasks))
	}
	if _, err := l.Get(1); err != ErrTaskNotFound {
		t.Error("removed task should be gone")
	}
	if err := l.Remove(1); err != ErrTaskNotFound {
		t.Errorf("Remove(missing) = %v, want ErrTaskNotFound", err)
	}
}
