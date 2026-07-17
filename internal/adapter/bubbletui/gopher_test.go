package bubbletui

import (
	"image"
	"image/color"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kendallowen/notebook/internal/todo"
	"github.com/muesli/termenv"
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

func TestHalfBlocksEmitsExactColors(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	// 1 wide x 2 tall: top pixel and bottom pixel are distinct known colors.
	img := image.NewRGBA(image.Rect(0, 0, 1, 2))
	img.Set(0, 0, color.RGBA{10, 20, 30, 255})    // top -> foreground
	img.Set(0, 1, color.RGBA{200, 150, 100, 255}) // bottom -> background
	out := halfBlocks(img)

	// hexOf's >>8 shift must yield the raw 8-bit components, and lipgloss must
	// render them as decimal truecolor escapes: fg=top, bg=bottom (not swapped).
	if !strings.Contains(out, "38;2;10;20;30") {
		t.Errorf("missing top-pixel foreground escape 38;2;10;20;30 in %q", out)
	}
	if !strings.Contains(out, "48;2;200;150;100") {
		t.Errorf("missing bottom-pixel background escape 48;2;200;150;100 in %q", out)
	}
	if !strings.Contains(out, "▀") {
		t.Errorf("missing half-block glyph in %q", out)
	}
}

func TestRenderGopherGuardReturnsSingleHintLine(t *testing.T) {
	out := renderGopher(1, 1)
	if lines := strings.Split(out, "\n"); len(lines) != 1 {
		t.Fatalf("got %d lines, want 1 (out=%q)", len(lines), out)
	}
	if !strings.Contains(out, "press any key to return") {
		t.Errorf("guard output missing hint text: %q", out)
	}
}

func TestGopherCtrlCQuits(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.mode = modeGopher
	cmd := send(m, key("ctrl+c"))
	if cmd == nil {
		t.Fatal("ctrl+c in gopher mode should return a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("ctrl+c command should produce tea.QuitMsg")
	}
}

func TestRenderGopherFillsViewport(t *testing.T) {
	out := renderGopher(80, 24)
	lines := strings.Split(out, "\n")
	if len(lines) != 24 {
		t.Fatalf("got %d lines, want 24", len(lines))
	}
	if !strings.Contains(out, "press any key to return") {
		t.Error("missing dismiss hint")
	}
}

func TestGopherTitleTriggersEasterEgg(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	typeStr(m, "gopher")
	m.updateForm(key("enter"))
	if m.mode != modeGopher {
		t.Fatalf("mode = %v, want modeGopher", m.mode)
	}
}

func TestNonGopherTitleDoesNotTrigger(t *testing.T) {
	for _, title := range []string{"Gopher", "gopher notes", "gophers"} {
		m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
		m.openAddTask()
		typeStr(m, title)
		m.updateForm(key("enter"))
		if m.mode != modeNormal {
			t.Errorf("title %q: mode = %v, want modeNormal", title, m.mode)
		}
	}
}

func TestGopherDismissedByAnyKey(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.focus = focusTasks
	m.mode = modeGopher
	m.Update(key("j"))
	if m.mode != modeNormal {
		t.Fatalf("mode = %v, want modeNormal", m.mode)
	}
	if m.focus != focusTasks {
		t.Errorf("focus = %v, want unchanged focusTasks", m.focus)
	}
}
