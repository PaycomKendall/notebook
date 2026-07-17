package bubbletui

import "testing"

func TestGopherImageDecodes(t *testing.T) {
	img := gopherImage()
	if img == nil {
		t.Fatal("gopherImage() returned nil")
	}
	b := img.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		t.Fatalf("bad bounds %v", b)
	}
}
