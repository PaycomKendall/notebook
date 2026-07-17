# Gopher Easter Egg Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When a user creates a page (task) titled exactly `gopher`, the Bubble Tea TUI takes over the full terminal to display a gopher photo rendered as truecolor half-blocks, dismissed by any key.

**Architecture:** A new `modeGopher` mode in the existing `internal/adapter/bubbletui` Model. The image is a downscaled JPEG embedded via `//go:embed`, decoded once, scaled to the viewport, and drawn as `▀` half-block cells (fg = upper pixel, bg = lower pixel). The add-task submit path sets the mode; the mode's key handler restores `modeNormal` on any key.

**Tech Stack:** Go 1.24, Bubble Tea v1.3.10, lipgloss v1.1.0, Go stdlib `image`/`image/jpeg`. No new module dependencies (scaling is hand-rolled nearest-neighbor).

## Global Constraints

- Module path: `github.com/kendallowen/notebook`.
- All new code lives in package `bubbletui` under `internal/adapter/bubbletui/`.
- No new go.mod dependencies — use stdlib + already-present lipgloss.
- Trigger match is exact and case-sensitive on the trimmed title: `strings.TrimSpace(title) == "gopher"`.
- Trigger fires only on the add-task path, never on edit.
- Any key (except `ctrl+c`, which quits) dismisses the egg and restores `modeNormal` with focus/selection unchanged.
- Run tests with: `go test ./internal/adapter/bubbletui/...`

---

### Task 1: Embed and decode the gopher image

**Files:**
- Create: `internal/adapter/bubbletui/gopher.jpg` (downscaled asset)
- Create: `internal/adapter/bubbletui/gopher.go`
- Test: `internal/adapter/bubbletui/gopher_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `func gopherImage() image.Image` — returns the decoded embedded image, or `nil` if decoding failed. Cached via `sync.Once`.

- [ ] **Step 1: Create the downscaled asset**

The source photo is at `/Users/kendallowen/.claude/image-cache/fdf37cf2-71d6-4df8-945b-1456a73b8241/2.jpeg`. Downscale its long edge to 400px into the package directory (macOS `sips`):

```bash
sips -Z 400 /Users/kendallowen/.claude/image-cache/fdf37cf2-71d6-4df8-945b-1456a73b8241/2.jpeg \
  --out internal/adapter/bubbletui/gopher.jpg
```

Verify it exists and is small:

```bash
ls -la internal/adapter/bubbletui/gopher.jpg
```

Expected: a JPEG on the order of tens of KB.

- [ ] **Step 2: Write the failing test**

Create `internal/adapter/bubbletui/gopher_test.go`:

```go
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
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/adapter/bubbletui/ -run TestGopherImageDecodes -v`
Expected: FAIL / build error — `undefined: gopherImage`.

- [ ] **Step 4: Write minimal implementation**

Create `internal/adapter/bubbletui/gopher.go`:

```go
package bubbletui

import (
	"bytes"
	_ "embed"
	"image"
	"image/jpeg"
	"sync"
)

//go:embed gopher.jpg
var gopherJPG []byte

var (
	gopherOnce sync.Once
	gopherImg  image.Image
)

// gopherImage decodes the embedded gopher photo once and caches it. Returns nil
// if decoding fails (should not happen with the committed asset).
func gopherImage() image.Image {
	gopherOnce.Do(func() {
		img, err := jpeg.Decode(bytes.NewReader(gopherJPG))
		if err == nil {
			gopherImg = img
		}
	})
	return gopherImg
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/adapter/bubbletui/ -run TestGopherImageDecodes -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/adapter/bubbletui/gopher.jpg internal/adapter/bubbletui/gopher.go internal/adapter/bubbletui/gopher_test.go
git commit -m "feat(bubbletui): embed and decode gopher easter-egg image"
```

---

### Task 2: Aspect-fit sizing and nearest-neighbor scaling

**Files:**
- Modify: `internal/adapter/bubbletui/gopher.go`
- Test: `internal/adapter/bubbletui/gopher_test.go`

**Interfaces:**
- Consumes: nothing from Task 1's functions (pure helpers).
- Produces:
  - `func fitDims(srcW, srcH, maxCols, maxRowsPx int) (w, h int)` — largest w×h that fits inside `maxCols`×`maxRowsPx` preserving aspect ratio; `h` is always even (half-block row pairs) and `w>=1`, `h>=2`.
  - `func scaleNearest(src image.Image, w, h int) *image.RGBA` — nearest-neighbor resample to exactly `w`×`h`.

- [ ] **Step 1: Write the failing tests**

Append to `internal/adapter/bubbletui/gopher_test.go`:

```go
import (
	"image"
	"testing"
)

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
```

Note: `gopher_test.go` already declares `package bubbletui` and imports `testing`; merge the `image` import into a single import block rather than duplicating.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/adapter/bubbletui/ -run 'TestFitDims|TestScaleNearest' -v`
Expected: FAIL / build error — `undefined: fitDims`, `undefined: scaleNearest`.

- [ ] **Step 3: Write minimal implementation**

Append to `internal/adapter/bubbletui/gopher.go` (add `"image"` to its import block — `image` is already imported from Task 1, so no change needed there):

```go
// fitDims returns the largest width/height (in pixels) that fits inside
// maxCols x maxRowsPx while preserving src aspect ratio. Height is forced even
// so it splits cleanly into half-block row pairs.
func fitDims(srcW, srcH, maxCols, maxRowsPx int) (w, h int) {
	if srcW <= 0 || srcH <= 0 {
		return 0, 0
	}
	s := float64(maxCols) / float64(srcW)
	if sh := float64(maxRowsPx) / float64(srcH); sh < s {
		s = sh
	}
	w = int(float64(srcW) * s)
	h = int(float64(srcH) * s)
	if w < 1 {
		w = 1
	}
	if h < 2 {
		h = 2
	}
	if h%2 == 1 {
		h--
	}
	return w, h
}

// scaleNearest resamples src to exactly w x h using nearest-neighbor sampling.
func scaleNearest(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sb := src.Bounds()
	for y := 0; y < h; y++ {
		sy := sb.Min.Y + y*sb.Dy()/h
		for x := 0; x < w; x++ {
			sx := sb.Min.X + x*sb.Dx()/w
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/adapter/bubbletui/ -run 'TestFitDims|TestScaleNearest' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/bubbletui/gopher.go internal/adapter/bubbletui/gopher_test.go
git commit -m "feat(bubbletui): aspect-fit sizing and nearest-neighbor scaling"
```

---

### Task 3: Half-block cell rendering

**Files:**
- Modify: `internal/adapter/bubbletui/gopher.go`
- Test: `internal/adapter/bubbletui/gopher_test.go`

**Interfaces:**
- Consumes: nothing from prior tasks (pure).
- Produces:
  - `func halfBlockCell(top, bottom color.Color) string` — one `▀` glyph styled with fg=top, bg=bottom (24-bit hex).
  - `func halfBlocks(img image.Image) string` — converts an image into lines of `▀` cells, pairing rows (row 0 over row 1, row 2 over row 3, …). Cell rows are joined with `\n`; odd trailing row is ignored.

- [ ] **Step 1: Write the failing test**

Append to `internal/adapter/bubbletui/gopher_test.go`:

```go
import "strings"

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
```

Merge the `"strings"` import into the existing import block.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/bubbletui/ -run TestHalfBlocksGridShape -v`
Expected: FAIL / build error — `undefined: halfBlocks`.

- [ ] **Step 3: Write minimal implementation**

Append to `internal/adapter/bubbletui/gopher.go` and add `"fmt"`, `"image/color"`, `"strings"`, and `"github.com/charmbracelet/lipgloss"` to its import block:

```go
// halfBlockCell renders one terminal cell as an upper half-block whose
// foreground is the top pixel and background is the bottom pixel.
func halfBlockCell(top, bottom color.Color) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(hexOf(top))).
		Background(lipgloss.Color(hexOf(bottom))).
		Render("▀")
}

func hexOf(c color.Color) string {
	r, g, b, _ := c.RGBA() // 16-bit per channel
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// halfBlocks converts img into rows of half-block cells. Each cell packs two
// vertical pixels; source rows are consumed in pairs. A dangling final row
// (odd height) is dropped.
func halfBlocks(img image.Image) string {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	var rows []string
	for y := 0; y+1 < h; y += 2 {
		var sb strings.Builder
		for x := 0; x < w; x++ {
			top := img.At(b.Min.X+x, b.Min.Y+y)
			bottom := img.At(b.Min.X+x, b.Min.Y+y+1)
			sb.WriteString(halfBlockCell(top, bottom))
		}
		rows = append(rows, sb.String())
	}
	return strings.Join(rows, "\n")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/bubbletui/ -run TestHalfBlocksGridShape -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/bubbletui/gopher.go internal/adapter/bubbletui/gopher_test.go
git commit -m "feat(bubbletui): render images as truecolor half-blocks"
```

---

### Task 4: `modeGopher` mode, full-screen render, and View case

**Files:**
- Modify: `internal/adapter/bubbletui/model.go:18-26` (mode enum)
- Modify: `internal/adapter/bubbletui/gopher.go` (add `renderGopher`)
- Modify: `internal/adapter/bubbletui/view.go:12-21` (View switch)
- Test: `internal/adapter/bubbletui/gopher_test.go`

**Interfaces:**
- Consumes: `gopherImage()`, `fitDims`, `scaleNearest`, `halfBlocks` from Tasks 1-3.
- Produces:
  - `modeGopher` — new value in the `mode` enum.
  - `func renderGopher(width, height int) string` — full-screen (width×height) string: the scaled gopher centered, with a `press any key to return` hint, centered via `lipgloss.Place`.
  - `func (m *Model) gopherView() string` — Model wrapper mirroring `normalView`'s paper-theme background handling.

- [ ] **Step 1: Add the mode enum value**

In `internal/adapter/bubbletui/model.go`, add `modeGopher` to the end of the mode const block:

```go
const (
	modeNormal mode = iota
	modeAddTask
	modeEditTask
	modeNewList
	modeRenameList
	modeMoveTask
	modeConfirm
	modeGopher
)
```

- [ ] **Step 2: Write the failing test**

Append to `internal/adapter/bubbletui/gopher_test.go`:

```go
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
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/adapter/bubbletui/ -run TestRenderGopherFillsViewport -v`
Expected: FAIL / build error — `undefined: renderGopher`.

- [ ] **Step 4: Implement `renderGopher`**

Append to `internal/adapter/bubbletui/gopher.go`:

```go
const gopherHint = "press any key to return"

// renderGopher produces a width x height full-screen view: the gopher photo
// scaled to fit (reserving one row for the hint), horizontally and vertically
// centered, with the dismiss hint beneath it.
func renderGopher(width, height int) string {
	if width < 2 || height < 3 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, gopherHint)
	}
	img := gopherImage()
	if img == nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, gopherHint)
	}
	b := img.Bounds()
	// height-1 text rows for the image (1 row reserved for the hint), each row
	// holds 2 vertical pixels.
	w, h := fitDims(b.Dx(), b.Dy(), width, (height-1)*2)
	art := halfBlocks(scaleNearest(img, w, h))
	block := lipgloss.JoinVertical(lipgloss.Center, art, gopherHint)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, block)
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/adapter/bubbletui/ -run TestRenderGopherFillsViewport -v`
Expected: PASS

- [ ] **Step 6: Add the View case and `gopherView`**

In `internal/adapter/bubbletui/view.go`, add a case to `View()`:

```go
func (m *Model) View() string {
	switch m.mode {
	case modeAddTask, modeEditTask, modeNewList, modeRenameList, modeMoveTask:
		return m.formView()
	case modeConfirm:
		return m.confirmView()
	case modeGopher:
		return m.gopherView()
	default:
		return m.normalView()
	}
}
```

Add `gopherView` to `view.go` (mirrors `normalView`'s paper-theme wrapping):

```go
func (m *Model) gopherView() string {
	content := renderGopher(m.width, m.height)
	if m.theme.bg != nil && m.width > 0 && m.height > 0 {
		return m.styles.page.Width(m.width).Height(m.height).Render(content)
	}
	return content
}
```

- [ ] **Step 7: Verify build and full package tests**

Run: `go build ./... && go test ./internal/adapter/bubbletui/...`
Expected: build succeeds; all tests PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/adapter/bubbletui/model.go internal/adapter/bubbletui/gopher.go internal/adapter/bubbletui/view.go internal/adapter/bubbletui/gopher_test.go
git commit -m "feat(bubbletui): modeGopher full-screen render and View case"
```

---

### Task 5: Trigger on add, and dismiss handler

**Files:**
- Modify: `internal/adapter/bubbletui/forms.go:164-176` (modeAddTask submit branch)
- Modify: `internal/adapter/bubbletui/update.go:14-23` (key routing) and add `updateGopher`
- Test: `internal/adapter/bubbletui/gopher_test.go`

**Interfaces:**
- Consumes: `modeGopher` (Task 4).
- Produces: `func (m *Model) updateGopher(msg tea.KeyMsg) (tea.Model, tea.Cmd)` — any key restores `modeNormal`; `ctrl+c` quits.

- [ ] **Step 1: Write the failing tests**

Append to `internal/adapter/bubbletui/gopher_test.go` (the `todo` package is already used by other tests in this package via `newTestModel`; this file needs `todo` imported — add `"github.com/kendallowen/notebook/internal/todo"` to its import block):

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/adapter/bubbletui/ -run 'TestGopher|TestNonGopher' -v`
Expected: FAIL — `TestGopherTitleTriggersEasterEgg` fails (mode is modeNormal, not modeGopher); `TestGopherDismissedByAnyKey` fails or panics (no `modeGopher` routing yet, falls to `updateForm`).

- [ ] **Step 3: Add the trigger in `submitForm`**

In `internal/adapter/bubbletui/forms.go`, replace the `modeAddTask` branch body (currently lines ~164-176) so `err` is captured and the mode flips after the reloads:

```go
	case modeAddTask:
		title := strings.TrimSpace(m.inputs[0].Value())
		if title == "" {
			return
		}
		tags := strings.Fields(m.inputs[1].Value())
		notes := m.inputs[2].Value()
		_, err := m.svc.AddTask(m.formList, title, tags, notes)
		if err != nil {
			m.status = err.Error()
		}
		m.closeForm()
		m.reloadLists()
		m.reloadTasks()
		if err == nil && title == "gopher" {
			m.mode = modeGopher // easter egg; closeForm reset mode to normal
		}
```

- [ ] **Step 4: Add the dismiss handler and route to it**

In `internal/adapter/bubbletui/update.go`, add a `modeGopher` case to the key switch in `Update`:

```go
	case tea.KeyMsg:
		switch m.mode {
		case modeConfirm:
			return m.updateConfirm(msg)
		case modeGopher:
			return m.updateGopher(msg)
		case modeNormal:
			return m.updateNormal(msg)
		default:
			return m.updateForm(msg)
		}
```

Add `updateGopher` to `update.go`:

```go
// updateGopher dismisses the gopher easter egg on any key (ctrl+c still quits),
// restoring the normal view with focus and selection untouched.
func (m *Model) updateGopher(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	m.mode = modeNormal
	return m, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/adapter/bubbletui/ -run 'TestGopher|TestNonGopher' -v`
Expected: PASS

- [ ] **Step 6: Run the full package test suite**

Run: `go test ./internal/adapter/bubbletui/...`
Expected: PASS (no regressions).

- [ ] **Step 7: Commit**

```bash
git add internal/adapter/bubbletui/forms.go internal/adapter/bubbletui/update.go internal/adapter/bubbletui/gopher_test.go
git commit -m "feat(bubbletui): trigger gopher egg on add, dismiss on any key"
```

---

### Task 6: Full build, suite, and manual verification

**Files:** none (verification only).

- [ ] **Step 1: Build and run the whole test suite**

Run: `go build ./... && go test ./...`
Expected: build succeeds; all packages PASS.

- [ ] **Step 2: Manual smoke test in a truecolor terminal**

Run: `go run ./cmd/nb`
Then: focus the Pages pane, press `a`, enter title `gopher`, submit. Expect the terminal to fill with the gopher photo and a `press any key to return` hint. Press any key and confirm you return to the three panes with the new `gopher` page present. Repeat with title `Gopher` and confirm the egg does NOT fire.

- [ ] **Step 3: Commit (if any incidental fixes were needed)**

```bash
git add -A
git commit -m "chore(bubbletui): verify gopher easter egg end-to-end"
```

---

## Self-Review

**Spec coverage:**
- Trigger (exact lowercase `gopher`, add-only, any folder) → Task 5, `submitForm` branch + `TestNonGopherTitleDoesNotTrigger`. ✓
- Full-screen takeover replacing panes → Task 4, `modeGopher` View case + `gopherView`. ✓
- Half-block truecolor rendering → Task 3 (`halfBlocks`/`halfBlockCell`) + Task 2 (scaling). ✓
- Embedded, self-contained asset → Task 1 (`//go:embed`). ✓
- Any-key dismiss, focus/selection preserved → Task 5, `updateGopher` + `TestGopherDismissedByAnyKey`. ✓
- `ctrl+c` still quits → Task 5, `updateGopher`. ✓
- Tests: converter, scaling, trigger, dismiss → Tasks 1-5. ✓
- Non-goals (no kitty/sixel, no edit trigger, no config flag) → respected; edit path untouched. ✓

**Placeholder scan:** No TBD/TODO/"handle edge cases"; every code step shows complete code. ✓

**Type consistency:** `gopherImage`, `fitDims`, `scaleNearest`, `halfBlocks`, `halfBlockCell`, `renderGopher`, `gopherView`, `updateGopher`, and `modeGopher` are used with identical names/signatures across the tasks that define and consume them. ✓
