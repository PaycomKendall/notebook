package bubbletui

import (
	"image"
	"strings"
	"testing"
)

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

func TestFitDimsPreservesAspectAndEvenHeight(t *testing.T) {
	cases := []struct {
		srcW, srcH, maxCols, maxRowsPx, wantW, wantH int
	}{
		{1000, 1000, 46, 46, 46, 46}, // square, height-bound tie
		{200, 100, 80, 40, 80, 40},   // wide, width-bound
		{100, 200, 80, 40, 20, 40},   // tall, height-bound
	}
	for _, c := range cases {
		w, h := fitDims(c.srcW, c.srcH, c.maxCols, c.maxRowsPx)
		if w != c.wantW || h != c.wantH {
			t.Errorf("fitDims(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
				c.srcW, c.srcH, c.maxCols, c.maxRowsPx, w, h, c.wantW, c.wantH)
		}
		if h%2 != 0 {
			t.Errorf("height %d must be even", h)
		}
	}
}

func TestScaleNearestOutputDimensions(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	dst := scaleNearest(src, 2, 2)
	if b := dst.Bounds(); b.Dx() != 2 || b.Dy() != 2 {
		t.Fatalf("scaled bounds = %v, want 2x2", b)
	}
}

func TestHalfBlocksGridShape(t *testing.T) {
	// 2 wide x 4 tall -> 2 cell rows of 2 cells each.
	img := image.NewRGBA(image.Rect(0, 0, 2, 4))
	out := halfBlocks(img)
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d cell rows, want 2 (out=%q)", len(lines), out)
	}
	if n := strings.Count(out, "▀"); n != 4 {
		t.Errorf("got %d half-block glyphs, want 4", n)
	}
}
