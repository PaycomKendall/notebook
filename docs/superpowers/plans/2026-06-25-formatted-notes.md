# Formatted Task Notes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow multi-line task notes with markdown styling in both TUIs, and make the Bubble Tea Detail pane look like a notebook page (widened, with header band, spiral gutter, margin rule, and ruled lines).

**Architecture:** A new pure `internal/markdown` package renders a markdown subset to ANSI-styled text using injected lipgloss styles. Both TUIs call it for the Detail pane (tview translates the ANSI to its own tags). The Bubble Tea TUI additionally wraps the rendered note in notebook-page decoration and widens the pane. Multi-line input uses native components: a `formField` interface over `textinput`/`textarea` in Bubble Tea, and `Form.AddTextArea` in tview.

**Tech Stack:** Go 1.24, lipgloss, bubbletea/bubbles (`textarea`), tview/tcell, cobra. No new module dependencies.

## Global Constraints

- Module path: `github.com/kendallowen/notebook`.
- No new entries in `go.mod` require block — use only already-present libraries (`lipgloss`, `bubbles/textarea`, `tview`).
- `todo.Task.Notes` stays a `string`; storage format is unchanged and backward compatible.
- Feature scope: multi-line input + markdown rendering go in **both** TUIs; notebook-page styling and the widened pane split are **Bubble Tea only**.
- No CLI changes. No markdown in the task-list pane.
- Every task ends green: `go test ./...` and `go vet ./...` pass before commit.
- Match existing house style: small focused files, table-driven tests, `gofmt`.

---

### Task 1: `internal/markdown` renderer package

**Files:**
- Create: `internal/markdown/markdown.go`
- Test: `internal/markdown/markdown_test.go`

**Interfaces:**
- Consumes: nothing (pure package; depends only on `lipgloss` + stdlib).
- Produces:
  - `type Styles struct { H1, H2, H3, Bold, Italic, Code, Bullet lipgloss.Style }`
  - `func Render(src string, width int, st Styles) string` — returns ANSI-styled text. `width <= 0` disables hard-wrapping.

- [ ] **Step 1: Write the failing test**

```go
// internal/markdown/markdown_test.go
package markdown

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// testStyles uses simple, detectable styling so assertions can look for the
// content text without depending on exact ANSI codes.
func testStyles() Styles {
	bold := lipgloss.NewStyle().Bold(true)
	return Styles{
		H1: bold, H2: bold, H3: bold,
		Bold: bold, Italic: lipgloss.NewStyle().Italic(true),
		Code: lipgloss.NewStyle().Reverse(true),
		Bullet: lipgloss.NewStyle(),
	}
}

func TestRenderPlainPassthrough(t *testing.T) {
	out := Render("just text", 0, testStyles())
	if !strings.Contains(out, "just text") {
		t.Fatalf("plain text dropped: %q", out)
	}
}

func TestRenderHeaderStripsMarker(t *testing.T) {
	out := Render("# Title", 0, testStyles())
	if strings.Contains(out, "#") {
		t.Errorf("header marker not stripped: %q", out)
	}
	if !strings.Contains(out, "Title") {
		t.Errorf("header text missing: %q", out)
	}
}

func TestRenderBullet(t *testing.T) {
	out := Render("- milk", 0, testStyles())
	if !strings.Contains(out, "•") {
		t.Errorf("bullet glyph missing: %q", out)
	}
	if !strings.Contains(out, "milk") {
		t.Errorf("bullet text missing: %q", out)
	}
}

func TestRenderNumbered(t *testing.T) {
	out := Render("1. first", 0, testStyles())
	if !strings.Contains(out, "1.") || !strings.Contains(out, "first") {
		t.Errorf("numbered item malformed: %q", out)
	}
}

func TestRenderInlineMarkersConsumed(t *testing.T) {
	out := Render("a **b** c `d` _e_", 0, testStyles())
	for _, marker := range []string{"**", "`", "_"} {
		if strings.Contains(out, marker) {
			t.Errorf("inline marker %q not consumed: %q", marker, out)
		}
	}
	for _, text := range []string{"a", "b", "c", "d", "e"} {
		if !strings.Contains(out, text) {
			t.Errorf("inline text %q missing: %q", text, out)
		}
	}
}

func TestRenderWrapsToWidth(t *testing.T) {
	out := Render("one two three four five six seven eight", 10, testStyles())
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > 10 {
			t.Errorf("line exceeds width 10: %q (w=%d)", line, lipgloss.Width(line))
		}
	}
	if !strings.Contains(out, "\n") {
		t.Errorf("expected wrapping to produce multiple lines: %q", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/markdown/`
Expected: FAIL — `undefined: Render` / package has no non-test files.

- [ ] **Step 3: Write the implementation**

```go
// internal/markdown/markdown.go
package markdown

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles carries the lipgloss styles a host applies to each markdown element.
type Styles struct {
	H1, H2, H3, Bold, Italic, Code, Bullet lipgloss.Style
}

var (
	reCode    = regexp.MustCompile("`([^`]+)`")
	reBold    = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reItalicA = regexp.MustCompile(`\*([^*]+)\*`)
	reItalicU = regexp.MustCompile(`_([^_]+)_`)
	reNumber  = regexp.MustCompile(`^(\d+)\.\s+(.*)$`)
)

// Render converts a markdown subset to ANSI-styled text. width <= 0 disables
// hard-wrapping (the host is responsible for wrapping).
func Render(src string, width int, st Styles) string {
	lines := strings.Split(src, "\n")
	out := make([]string, 0, len(lines))
	for _, raw := range lines {
		styled := renderBlock(raw, st)
		if width > 0 {
			styled = lipgloss.NewStyle().Width(width).Render(styled)
		}
		out = append(out, styled)
	}
	return strings.Join(out, "\n")
}

// renderBlock styles a single source line as a header, list item, or paragraph,
// then applies inline styling to its text.
func renderBlock(line string, st Styles) string {
	switch {
	case strings.HasPrefix(line, "### "):
		return st.H3.Render(inline(strings.TrimPrefix(line, "### "), st))
	case strings.HasPrefix(line, "## "):
		return st.H2.Render(inline(strings.TrimPrefix(line, "## "), st))
	case strings.HasPrefix(line, "# "):
		return st.H1.Render(inline(strings.TrimPrefix(line, "# "), st))
	case strings.HasPrefix(line, "- "), strings.HasPrefix(line, "* "):
		return st.Bullet.Render("•") + " " + inline(line[2:], st)
	}
	if m := reNumber.FindStringSubmatch(line); m != nil {
		return m[1] + ". " + inline(m[2], st)
	}
	return inline(line, st)
}

// inline applies code, bold, then italic styling. Code is processed first so
// markers inside backticks are left alone; bold before italic so ** is consumed
// before single-* italic. Injected ANSI never contains *, _, or ` so later
// passes cannot be confused by it.
func inline(s string, st Styles) string {
	s = reCode.ReplaceAllStringFunc(s, func(m string) string {
		return st.Code.Render(reCode.FindStringSubmatch(m)[1])
	})
	s = reBold.ReplaceAllStringFunc(s, func(m string) string {
		return st.Bold.Render(reBold.FindStringSubmatch(m)[1])
	})
	s = reItalicA.ReplaceAllStringFunc(s, func(m string) string {
		return st.Italic.Render(reItalicA.FindStringSubmatch(m)[1])
	})
	s = reItalicU.ReplaceAllStringFunc(s, func(m string) string {
		return st.Italic.Render(reItalicU.FindStringSubmatch(m)[1])
	})
	return s
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/markdown/ && go vet ./internal/markdown/`
Expected: PASS (all tests), vet clean.

- [ ] **Step 5: Commit**

```bash
git add internal/markdown/
git commit -m "feat(markdown): add subset markdown-to-ANSI renderer"
```

---

### Task 2: Render markdown in the tview Detail pane

**Files:**
- Modify: `internal/adapter/tui/app.go` (`refreshDetail`, ~line 176-190; add an import)
- Test: `internal/adapter/tui/app_test.go` (add a test)

**Interfaces:**
- Consumes: `markdown.Render`, `markdown.Styles` from Task 1.
- Produces: nothing new for later tasks (tview Detail behavior only).

The Detail `TextView` already has `SetDynamicColors(true)` (`app.go:68`). We render the note through markdown, translate the ANSI to tview tags with `tview.TranslateANSI`, and pass `width = 0` so tview's own wrapping applies.

- [ ] **Step 1: Write the failing test**

The verified helper is `func newTestApp(t *testing.T) (*App, *todo.Service)`
(`app_test.go:11`); it seeds via the returned service, and `AddTask` auto-creates
the list. `app.detail` is created in `buildUI`, so call `buildUI` +
`refreshLists` + `refreshTasks` before `refreshDetail` (same order as
`TestSkeletonRendersListsAndTasks`). Add `"strings"` to the test file's imports.

```go
// internal/adapter/tui/app_test.go — add this test.
func TestDetailRendersMarkdownNote(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("inbox", "task", nil, "# Heading\n- item"); err != nil {
		t.Fatal(err)
	}
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.refreshDetail()

	got := app.detail.GetText(true) // true = strip color tags
	if strings.Contains(got, "# Heading") {
		t.Errorf("header marker should be stripped, got: %q", got)
	}
	if !strings.Contains(got, "Heading") || !strings.Contains(got, "•") {
		t.Errorf("rendered note missing content/bullet: %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/tui/ -run TestDetailRendersMarkdownNote`
Expected: FAIL — `GetText` still contains `# Heading` and no `•`.

- [ ] **Step 3: Implement**

Add the import in `internal/adapter/tui/app.go`:

```go
	"github.com/kendallowen/notebook/internal/markdown"
```

Add a package-level default style set (tview has no theme system, so use fixed
styles) near the top of `app.go`, after the imports:

```go
// detailMD is the markdown styling used in the tview Detail pane. tview has no
// theme palette, so these are fixed; markdown.Render emits ANSI which
// tview.TranslateANSI converts to tview's color tags.
var detailMD = markdown.Styles{
	H1:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
	H2:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
	H3:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
	Bold:   lipgloss.NewStyle().Bold(true),
	Italic: lipgloss.NewStyle().Italic(true),
	Code:   lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
	Bullet: lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
}
```

Add the lipgloss import to `app.go` if absent:

```go
	"github.com/charmbracelet/lipgloss"
```

Replace the notes line in `refreshDetail` (`app.go:188`):

```go
	// before:
	fmt.Fprintf(&b, "Notes:\n%s\n", t.Notes)
	// after:
	fmt.Fprintf(&b, "Notes:\n%s\n", tview.TranslateANSI(markdown.Render(t.Notes, 0, detailMD)))
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/adapter/tui/ && go vet ./internal/adapter/tui/`
Expected: PASS, vet clean.

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/tui/app.go internal/adapter/tui/app_test.go
git commit -m "feat(tui): render markdown in the Detail pane"
```

---

### Task 3: Bubble Tea `formField` abstraction (behavior-preserving refactor)

**Files:**
- Modify: `internal/adapter/bubbletui/model.go` (`inputs` field type)
- Modify: `internal/adapter/bubbletui/forms.go` (`newInput`, `refocusInputs`, `submitForm`, `updateForm`, `formView`, the `open*` constructors)
- Test: existing `internal/adapter/bubbletui/forms_test.go` / `update_test.go` must still pass.

**Interfaces:**
- Consumes: nothing new.
- Produces:
  - `type formField interface { Focus() tea.Cmd; Blur(); Update(tea.Msg) tea.Cmd; View() string; Value() string; SetValue(string); label() string; multiline() bool }`
  - `func newInput(label, value string) formField` (now returns the interface, wrapping `textinput`)
  - `m.inputs` is now `[]formField`.

This task introduces the interface and a single-line implementation, leaving the
form behaving exactly as today (Notes is still single-line). The textarea
arrives in Task 4.

- [ ] **Step 1: Add the interface + textinput wrapper**

Create the wrapper at the top of `internal/adapter/bubbletui/forms.go` (replace
the existing `newInput`):

```go
// formField is a single field in a modal form. It abstracts over single-line
// (textinput) and multi-line (textarea) widgets so the form loop stays generic.
type formField interface {
	Focus() tea.Cmd
	Blur()
	Update(tea.Msg) tea.Cmd
	View() string
	Value() string
	SetValue(string)
	label() string  // a heading printed above the widget; "" if the widget shows its own
	multiline() bool
}

// lineField wraps a single-line textinput. Its label lives in the input's
// Prompt, so label() returns "".
type lineField struct{ ti textinput.Model }

func (f *lineField) Focus() tea.Cmd { return f.ti.Focus() }
func (f *lineField) Blur()          { f.ti.Blur() }
func (f *lineField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.ti, cmd = f.ti.Update(msg)
	return cmd
}
func (f *lineField) View() string     { return f.ti.View() }
func (f *lineField) Value() string    { return f.ti.Value() }
func (f *lineField) SetValue(s string) { f.ti.SetValue(s) }
func (f *lineField) label() string    { return "" }
func (f *lineField) multiline() bool  { return false }

func newInput(label, value string) formField {
	ti := textinput.New()
	ti.Prompt = label + ": "
	ti.SetValue(value)
	ti.CharLimit = 200
	ti.Width = 36
	return &lineField{ti: ti}
}
```

- [ ] **Step 2: Update the model field type**

In `internal/adapter/bubbletui/model.go`, change the field:

```go
	// before:
	inputs      []textinput.Model
	// after:
	inputs      []formField
```

If `textinput` is no longer referenced in `model.go`, remove its import (run
`go build ./...` to find out).

- [ ] **Step 3: Update `forms.go` call sites**

The `open*` functions build `[]textinput.Model{...}`. Change each to
`[]formField{...}` (the `newInput(...)` calls already return `formField` now).
For example in `openAddTask`:

```go
	m.inputs = []formField{
		newInput("Title", ""),
		newInput("Tags (space-separated)", ""),
		newInput("Notes", ""),
	}
```

Apply the same `[]textinput.Model{` → `[]formField{` change in `openEditTask`,
`openNewList`, `openRenameList`, and `openMoveTask`.

Update `refocusInputs` to use the interface methods:

```go
func (m *Model) refocusInputs() {
	for i := range m.inputs {
		if i == m.formField {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}
```

Update `updateForm`'s delegation tail (the last two lines) to use the new
`Update` signature:

```go
	// before:
	var cmd tea.Cmd
	m.inputs[m.formField], cmd = m.inputs[m.formField].Update(msg)
	return m, cmd
	// after:
	cmd := m.inputs[m.formField].Update(msg)
	return m, cmd
```

Update `formView` to print labels for fields that have them:

```go
func (m *Model) formView() string {
	var b strings.Builder
	b.WriteString(m.styles.title.Render(m.formTitle()) + "\n\n")
	for i := range m.inputs {
		if lbl := m.inputs[i].label(); lbl != "" {
			b.WriteString(m.styles.dim.Render(lbl) + "\n")
		}
		b.WriteString(m.inputs[i].View() + "\n")
	}
	b.WriteString("\n" + m.styles.dim.Render("tab/↑↓: move · enter: submit · esc: cancel"))
	if m.status != "" {
		b.WriteString("\n" + m.styles.warn.Render(m.status))
	}
	return m.styles.modal.Render(b.String())
}
```

`submitForm` already calls `m.inputs[i].Value()` — no change needed.

- [ ] **Step 4: Run the existing tests to verify they still pass**

Run: `go test ./internal/adapter/bubbletui/ && go vet ./internal/adapter/bubbletui/`
Expected: PASS — the refactor is behavior-preserving (Notes is still single-line; Enter still submits).

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/bubbletui/model.go internal/adapter/bubbletui/forms.go
git commit -m "refactor(bubbletui): abstract form fields behind formField interface"
```

---

### Task 4: Bubble Tea Notes as a multi-line textarea

**Files:**
- Modify: `internal/adapter/bubbletui/forms.go` (add `areaField`, `newArea`, use it for Notes, update `updateForm` key rules and the hint)
- Test: `internal/adapter/bubbletui/forms_test.go` (add tests)

**Interfaces:**
- Consumes: `formField` from Task 3.
- Produces:
  - `func newArea(label, value string) formField` wrapping `textarea.Model` (`multiline() == true`).

Key rules in the form: **Enter** submits only when the focused field is
single-line; **Ctrl+S** always submits; **↑/↓** delegate to the field when it is
multi-line (cursor movement) and otherwise move between fields; **Tab/Shift-Tab**
always move between fields.

- [ ] **Step 1: Write the failing tests**

```go
// internal/adapter/bubbletui/forms_test.go — add these.

func TestNotesFieldIsMultiline(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	// Notes is the 3rd field (index 2).
	if !m.inputs[2].multiline() {
		t.Fatalf("Notes field should be multiline")
	}
}

func TestEnterInNotesInsertsNewlineNotSubmit(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	typeStr(m, "title")
	m.formField = 2 // Notes
	m.refocusInputs()
	typeStr(m, "line one")
	m.updateForm(key("enter")) // should NOT submit; should add a newline
	if m.mode != modeAddTask {
		t.Fatalf("enter in multiline Notes must not submit; mode=%v", m.mode)
	}
	typeStr(m, "line two")
	if !strings.Contains(m.inputs[2].Value(), "\n") {
		t.Errorf("notes should contain a newline, got %q", m.inputs[2].Value())
	}
}

func TestCtrlSSubmitsFromNotes(t *testing.T) {
	m, svc := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.openAddTask()
	typeStr(m, "buy milk")
	m.formField = 2
	m.refocusInputs()
	typeStr(m, "2%")
	m.updateForm(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.mode != modeNormal {
		t.Fatalf("ctrl+s should submit; mode=%v", m.mode)
	}
	l, _ := svc.GetList("work")
	if len(l.Tasks) != 1 || l.Tasks[0].Notes != "2%" {
		t.Fatalf("task not added with notes: %+v", l.Tasks)
	}
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/adapter/bubbletui/ -run 'Notes|CtrlS'`
Expected: FAIL — Notes is still single-line; `multiline()` is false; Ctrl+S unhandled.

- [ ] **Step 3: Add the textarea wrapper**

Add to `internal/adapter/bubbletui/forms.go` (add the textarea import):

```go
	"github.com/charmbracelet/bubbles/textarea"
```

```go
// areaField wraps a multi-line textarea. Its label is printed above it by
// formView (textarea has no label/Prompt concept like textinput).
type areaField struct {
	ta  textarea.Model
	lbl string
}

func (f *areaField) Focus() tea.Cmd { return f.ta.Focus() }
func (f *areaField) Blur()          { f.ta.Blur() }
func (f *areaField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.ta, cmd = f.ta.Update(msg)
	return cmd
}
func (f *areaField) View() string      { return f.ta.View() }
func (f *areaField) Value() string     { return f.ta.Value() }
func (f *areaField) SetValue(s string) { f.ta.SetValue(s) }
func (f *areaField) label() string     { return f.lbl }
func (f *areaField) multiline() bool   { return true }

func newArea(label, value string) formField {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.CharLimit = 2000
	ta.SetWidth(38)
	ta.SetHeight(5)
	ta.SetValue(value)
	return &areaField{ta: ta, lbl: label}
}
```

- [ ] **Step 4: Use the textarea for Notes**

In `openAddTask`, change the Notes field:

```go
	m.inputs = []formField{
		newInput("Title", ""),
		newInput("Tags (space-separated)", ""),
		newArea("Notes", ""),
	}
```

In `openEditTask`, change the Notes field:

```go
	m.inputs = []formField{
		newInput("Title", t.Title),
		newArea("Notes", t.Notes),
	}
```

- [ ] **Step 5: Update `updateForm` key rules**

Replace the body of `updateForm` with:

```go
func (m *Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cur := m.inputs[m.formField]
	switch msg.Type {
	case tea.KeyEsc:
		m.closeForm()
		return m, nil
	case tea.KeyCtrlS:
		m.submitForm()
		return m, nil
	case tea.KeyEnter:
		if !cur.multiline() {
			m.submitForm()
			return m, nil
		}
		// multiline: fall through to delegate (inserts a newline)
	case tea.KeyTab:
		m.formField = (m.formField + 1) % len(m.inputs)
		m.refocusInputs()
		return m, nil
	case tea.KeyShiftTab:
		m.formField = (m.formField - 1 + len(m.inputs)) % len(m.inputs)
		m.refocusInputs()
		return m, nil
	case tea.KeyDown, tea.KeyUp:
		if !cur.multiline() {
			step := 1
			if msg.Type == tea.KeyUp {
				step = -1
			}
			m.formField = (m.formField + step + len(m.inputs)) % len(m.inputs)
			m.refocusInputs()
			return m, nil
		}
		// multiline: fall through to delegate (cursor moves between lines)
	}
	cmd := cur.Update(msg)
	return m, cmd
}
```

Update the hint in `formView`:

```go
	b.WriteString("\n" + m.styles.dim.Render("tab: move · ctrl+s: submit · esc: cancel"))
```

Note: the old code matched on `msg.String()`; this version matches on
`msg.Type`, which is why `key("tab")`/`key("enter")`/`key("esc")` in the tests
(built from `tea.KeyTab`/`tea.KeyEnter`/`tea.KeyEscape`) keep working.

- [ ] **Step 6: Run tests**

Run: `go test ./internal/adapter/bubbletui/ && go vet ./internal/adapter/bubbletui/`
Expected: PASS, including the existing `TestAddTaskFormSubmits` (Enter on the
single-line Tags field still submits).

- [ ] **Step 7: Commit**

```bash
git add internal/adapter/bubbletui/forms.go internal/adapter/bubbletui/forms_test.go
git commit -m "feat(bubbletui): multi-line Notes textarea with ctrl+s submit"
```

---

### Task 5: Bubble Tea Detail pane — markdown + notebook page

**Files:**
- Modify: `internal/adapter/bubbletui/view.go` (`renderDetail`; add helpers; add imports)
- Modify: `internal/adapter/bubbletui/styles.go` (add a `markdown.Styles` builder)
- Test: `internal/adapter/bubbletui/view_test.go` (add tests)

**Interfaces:**
- Consumes: `markdown.Render`, `markdown.Styles`, `m.paneWidths`, `m.paneHeight`, `m.styles`.
- Produces: `func (m *Model) mdStyles() markdown.Styles`; constants `ndGutter`, `ndMargin`.

Decoration (approved option **D**): a `NOTEBOOK` header band + separator, then
each page row is `gutter + margin + ruled(text)`, filled to the pane height so
ruled lines continue past the note.

> **Known caveat (document, don't block):** lipgloss underline applied to a line
> that already contains inline ANSI (bold/code spans) may not underline the
> styled span on every terminal. The dominant visual — ruled lines extending
> down the empty part of the page — is unaffected because filler rows are plain.
> This is acceptable for v1; a visual check is included in Step 6.

- [ ] **Step 1: Add the markdown style builder**

In `internal/adapter/bubbletui/styles.go`, add (and add the imports
`"github.com/charmbracelet/lipgloss"` is already present;
add `"github.com/kendallowen/notebook/internal/markdown"`):

```go
// mdStyles maps the active theme to markdown element styles for the Detail pane.
func (m *Model) mdStyles() markdown.Styles {
	return markdown.Styles{
		H1:     lipgloss.NewStyle().Bold(true).Foreground(m.themeAccent()),
		H2:     lipgloss.NewStyle().Bold(true).Foreground(m.themeAccent()),
		H3:     lipgloss.NewStyle().Bold(true).Foreground(m.themeAccent()),
		Bold:   lipgloss.NewStyle().Bold(true),
		Italic: lipgloss.NewStyle().Italic(true),
		Code:   m.styles.tag,
		Bullet: m.styles.tag,
	}
}
```

`Styles` only stores computed lipgloss styles, not raw colors, so add a small
accessor. In `model.go`, store the theme on the model: change `New` to keep it.
Add to the `Model` struct:

```go
	theme  Theme
```

and in `New`:

```go
	m := &Model{svc: svc, width: 90, height: 24, focus: focusTasks, theme: theme, styles: theme.styles()}
```

Then add to `styles.go`:

```go
func (m *Model) themeAccent() lipgloss.TerminalColor { return m.theme.accent }
```

- [ ] **Step 2: Write the failing tests**

```go
// internal/adapter/bubbletui/view_test.go — add these.
// (Reuse the existing newTestModel helper used by other view tests.)

func TestDetailShowsNotebookChrome(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "task", nil, "# Plan\n- step one")
	})
	m.focus = focusDetail
	out := m.renderDetail()
	if !strings.Contains(out, "◦") {
		t.Errorf("expected spiral binding gutter, got:\n%s", out)
	}
	if !strings.Contains(out, "N O T E B O O K") {
		t.Errorf("expected header band, got:\n%s", out)
	}
	if !strings.Contains(out, "•") {
		t.Errorf("expected rendered bullet, got:\n%s", out)
	}
	if strings.Contains(out, "# Plan") {
		t.Errorf("markdown header marker should be stripped, got:\n%s", out)
	}
}

func TestDetailEmptyNoteStillLined(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) {
		_, _ = s.AddTask("work", "task", nil, "")
	})
	m.focus = focusDetail
	out := m.renderDetail()
	// Gutter glyph appears on filler rows even with no note text.
	if strings.Count(out, "◦") < 2 {
		t.Errorf("empty note should still render lined page, got:\n%s", out)
	}
}
```

- [ ] **Step 3: Run to verify failure**

Run: `go test ./internal/adapter/bubbletui/ -run 'Notebook|Lined'`
Expected: FAIL — current `renderDetail` has no gutter/band.

- [ ] **Step 4: Rewrite `renderDetail` + add helpers**

In `internal/adapter/bubbletui/view.go`, add the import
`"github.com/kendallowen/notebook/internal/markdown"`, and replace
`renderDetail` (lines ~91-104) with:

```go
const (
	ndGutter = "◦ " // spiral binding hole + space
	ndMargin = "│ " // margin rule + space
)

func (m *Model) renderDetail() string {
	_, _, dw := m.paneWidths()
	focused := m.focus == focusDetail
	// inner content width: pane width minus paneStyle's horizontal padding (1+1)
	// minus the gutter and margin prefixes.
	contentW := dw - 2 - lipgloss.Width(ndGutter) - lipgloss.Width(ndMargin)
	if contentW < 8 {
		contentW = 8
	}

	var lines []string
	if t := m.selectedTask(); t != nil {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render(t.Title))
		if len(t.Tags) > 0 {
			lines = append(lines, m.styles.tag.Render("#"+strings.Join(t.Tags, " #")))
		}
		lines = append(lines, "")
		lines = append(lines, strings.Split(markdown.Render(t.Notes, contentW, m.mdStyles()), "\n")...)
	}

	body := m.titleFor("Detail", focused) + "\n\n" + m.notebookPage(lines, contentW)
	return m.paneStyle(focused).Width(dw).Height(m.paneHeight()).Render(strings.TrimRight(body, "\n"))
}

// notebookPage decorates content lines as a notebook page: a header band, a
// separator rule, then guttered + margined + ruled rows filling the pane height.
func (m *Model) notebookPage(lines []string, contentW int) string {
	gutter := m.styles.dim.Render(ndGutter)
	margin := m.styles.tag.Render(ndMargin)
	rule := lipgloss.NewStyle().Underline(true).Foreground(m.theme.subtle)

	var b strings.Builder
	b.WriteString(gutter + margin + m.styles.dim.Render("N O T E B O O K") + "\n")
	b.WriteString(gutter + margin + m.styles.dim.Render(strings.Repeat("─", contentW)) + "\n")

	// rows = pane height minus the Detail title (1), the blank line (1),
	// the header band (1) and the separator (1).
	rows := m.paneHeight() - 4
	if rows < 1 {
		rows = 1
	}
	for i := 0; i < rows; i++ {
		text := ""
		if i < len(lines) {
			text = lines[i]
		}
		// pad to contentW so the underline spans the full page width.
		pad := contentW - lipgloss.Width(text)
		if pad < 0 {
			pad = 0
		}
		b.WriteString(gutter + margin + rule.Render(text+strings.Repeat(" ", pad)) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/adapter/bubbletui/ && go vet ./internal/adapter/bubbletui/`
Expected: PASS.

- [ ] **Step 6: Visual check (manual)**

Run: `go run ./cmd/nb tui --engine bubble --theme nord`
Add/select a task with a note like `# Plan`, a `- bullet`, `**bold**`, and
`` `code` ``. Confirm the Detail pane shows the header band, the `◦` gutter, the
colored margin rule, ruled lines extending down the page, and styled markdown.
Press `q` to quit. (No automated assertion — this is the look-and-feel gate.)

- [ ] **Step 7: Commit**

```bash
git add internal/adapter/bubbletui/view.go internal/adapter/bubbletui/styles.go internal/adapter/bubbletui/model.go internal/adapter/bubbletui/view_test.go
git commit -m "feat(bubbletui): render notes as a markdown notebook page"
```

---

### Task 6: Widen the Bubble Tea Detail pane (40/60 split)

**Files:**
- Modify: `internal/adapter/bubbletui/view.go` (`paneWidths`, ~line 22-31)
- Test: `internal/adapter/bubbletui/view_test.go` (add a test)

**Interfaces:**
- Consumes: nothing new.
- Produces: nothing new.

- [ ] **Step 1: Write the failing test**

```go
// internal/adapter/bubbletui/view_test.go — add this.
func TestDetailPaneWiderThanTasks(t *testing.T) {
	m, _ := newTestModel(t, func(s *todo.Service) { _ = s.CreateList("work") })
	m.width = 120
	m.height = 30
	_, tasks, detail := m.paneWidths()
	if detail <= tasks {
		t.Errorf("detail (%d) should be wider than tasks (%d)", detail, tasks)
	}
}
```

- [ ] **Step 2: Run to verify failure**

Run: `go test ./internal/adapter/bubbletui/ -run TestDetailPaneWiderThanTasks`
Expected: FAIL — currently tasks (3/5) > detail (2/5).

- [ ] **Step 3: Flip the split**

In `paneWidths`, change the two computation lines:

```go
	// before:
	tasks = avail * 3 / 5
	detail = avail - tasks
	// after:
	tasks = avail * 2 / 5
	detail = avail - tasks
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/adapter/bubbletui/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/bubbletui/view.go internal/adapter/bubbletui/view_test.go
git commit -m "feat(bubbletui): widen Detail pane to a 40/60 split"
```

---

### Task 7: Multi-line Notes input in the tview forms

**Files:**
- Modify: `internal/adapter/tui/app.go` (App struct: add `lastForm` field)
- Modify: `internal/adapter/tui/forms.go` (`showModalForm` records the form; `addTaskForm`, `editTaskForm`; the hint)
- Test: `internal/adapter/tui/forms_test.go` (add a test)

**Interfaces:**
- Consumes: `tview.Form.AddTextArea`, `Form.GetFormItemByLabel`, `Form.GetButton`.
- Produces: `a.lastForm *tview.Form` — a test seam holding the most recently
  shown modal form.

tview forms submit via the Add/Save **button**, so there is no Enter conflict —
this is a near drop-in swap of the Notes `AddInputField` for `AddTextArea`.

There is no `Application.GetRoot`, and the form delegates focus to its first
item, so `GetFocus` does not return the `*tview.Form`. Add a one-field seam:
`showModalForm` records the form it shows, and the test reads it back.

- [ ] **Step 1: Add the `lastForm` seam**

In `internal/adapter/tui/app.go`, add a field to the `App` struct (next to the
existing `detail`/`footer` fields):

```go
	lastForm *tview.Form // most recently shown modal form (test seam)
```

In `internal/adapter/tui/forms.go`, record it at the top of `showModalForm`:

```go
func (a *App) showModalForm(form *tview.Form, title string) {
	a.lastForm = form
	form.SetBorder(true).SetTitle(title)
	// ...rest unchanged...
```

- [ ] **Step 2: Write the failing test**

The verified helper is `func newTestApp(t *testing.T) (*App, *todo.Service)`.
`tcell` is already imported by tests in this package.

```go
// internal/adapter/tui/forms_test.go — add this.
// Verifies a multi-line note round-trips through the edit form's Save handler.
func TestEditFormAcceptsMultilineNotes(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("inbox", "task", nil, "old"); err != nil {
		t.Fatal(err)
	}
	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.editTaskForm() // builds the form and records app.lastForm

	item := app.lastForm.GetFormItemByLabel("Notes")
	ta, ok := item.(*tview.TextArea)
	if !ok {
		t.Fatalf("Notes item is %T, want *tview.TextArea", item)
	}
	ta.SetText("line one\nline two", true)

	// Trigger the Save button (index 0) via its input handler.
	app.lastForm.GetButton(0).InputHandler()(
		tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone),
		func(tview.Primitive) {},
	)

	l, _ := svc.GetList("inbox")
	if l.Tasks[0].Notes != "line one\nline two" {
		t.Errorf("notes = %q, want multi-line", l.Tasks[0].Notes)
	}
}
```

Add the `tview` import to `forms_test.go` if not already present
(`"github.com/rivo/tview"`).

- [ ] **Step 3: Run to verify failure**

Run: `go test ./internal/adapter/tui/ -run TestEditFormAcceptsMultilineNotes`
Expected: FAIL — Notes is an `*tview.InputField`, so the type assertion fails.

- [ ] **Step 4: Swap to `AddTextArea`**

In `addTaskForm` (`internal/adapter/tui/forms.go:75`), replace the Notes line:

```go
	// before:
	AddInputField("Notes", "", 40, nil, func(s string) { notes = s })
	// after:
	AddTextArea("Notes", "", 40, 6, 0, func(s string) { notes = s })
```

In `editTaskForm` (`forms.go:99`), replace:

```go
	// before:
	AddInputField("Notes", notes, 40, nil, func(s string) { notes = s })
	// after:
	AddTextArea("Notes", notes, 40, 6, 0, func(s string) { notes = s })
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/adapter/tui/ && go vet ./internal/adapter/tui/`
Expected: PASS.

- [ ] **Step 6: Verify Tab navigation in a real terminal**

Run: `go run ./cmd/nb tui` (tview default). Press `a` to add a task. Confirm
`Tab`/`Shift-Tab` still move between Title → Tags → Notes → buttons while the
Notes area is focused, and that `Enter` inside Notes inserts a newline.

**If Tab is captured by the text area** (inserts a tab instead of moving focus):
add a form-level input capture in `showModalForm` that remaps Tab/Shift-Tab to
`form.SetFocus(next/prev)` when the focused item is a `*tview.TextArea`. Keep
this fallback minimal and only add it if the manual check shows it's needed.

- [ ] **Step 7: Update the form hint**

Only if Enter-in-Notes behavior changed the meaning of the hint. The current
hint (`forms.go:12`) reads `Enter: select`. Update to reflect text areas:

```go
const formHint = " Tab/↑↓: move  ·  Enter: newline in Notes  ·  Esc: cancel  (use buttons to submit)"
```

- [ ] **Step 8: Commit**

```bash
git add internal/adapter/tui/app.go internal/adapter/tui/forms.go internal/adapter/tui/forms_test.go
git commit -m "feat(tui): multi-line Notes text area in add/edit forms"
```

---

## Self-Review

**Spec coverage:**
- Multi-line input — Bubble Tea (Task 4), tview (Task 7). ✅
- Markdown rendering — shared renderer (Task 1), tview (Task 2), Bubble Tea (Task 5). ✅
- Notebook-page styling, Bubble Tea only (Task 5). ✅
- Widened Detail pane, Bubble Tea only (Task 6). ✅
- Storage unchanged / backward compatible — no `Task`/store changes in any task. ✅
- CLI unchanged — no `cmd/` or `internal/adapter/cli` task. ✅
- No new dependencies — only `textarea` (in `bubbles`, already required) and `markdown` (internal). ✅

**Type consistency:** `formField` interface (Task 3) is consumed unchanged by Task 4; `markdown.Styles`/`markdown.Render` (Task 1) are consumed by Tasks 2 and 5 with matching signatures; `mdStyles()`, `ndGutter`, `ndMargin` defined and used within Task 5.

**Placeholder scan:** No TBD/TODO; every code step shows complete code. The two
manual-verification steps (Task 5 Step 6, Task 7 Step 6) are intentional
look-and-feel / terminal-behavior gates for a visual feature, each with a
concrete fallback, not deferred work.

**Risk notes carried into execution:**
- Task 5: underline-over-styled-text caveat documented; filler rows guarantee the lined look regardless.
- Task 7: tview Tab-capture fallback specified; `lastForm` seam avoids the
  non-existent `Application.GetRoot` / unreliable `GetFocus` for grabbing the form.
- Verified against the real code: `newTestApp(t)` returns `(*App, *todo.Service)`
  and seeds via the service (`AddTask` auto-creates lists); bubbletui's
  `newTestModel(t, seedFn)` takes a seed callback. Tests match these shapes.
