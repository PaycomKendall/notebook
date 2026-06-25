# Formatted task notes — design

**Date:** 2026-06-25
**Status:** Approved

## Goal

Allow richer task notes in both TUIs:

1. **Multi-line entry** — today the add/edit Notes field is a single-line input.
   Make it a true multi-line text area so notes can hold paragraphs and line
   breaks.
2. **Markdown styling in the Detail pane** — render a useful subset of markdown
   (headers, bold, italic, inline code, bullet/numbered lists) instead of raw
   text, themed to match the active TUI theme.

## Decisions

- **Renderer:** a custom subset renderer in a new shared package. Chosen over
  `glamour` because the Detail pane is a narrow column in a three-pane layout
  (width/wrap control matters) and because glamour's own color themes would
  clash with the existing `nord`/`dracula`/`gruvbox`/`mono`/`default` themes.
  Zero new dependencies; fully unit-testable.
- **Scope:** both TUIs (tview default + Bubble Tea), matching the established
  feature-parity pattern. The CLI is unchanged: `-n` already accepts any string
  (newlines pass through), and markdown rendering is a display-only concern.
- **Storage:** unchanged. `todo.Task.Notes` stays a `string`, now permitted to
  contain newlines and markdown. JSON already round-trips this, so the change is
  fully backward compatible — existing single-line notes render unchanged.

## Components

### 1. `internal/markdown` (new package)

A pure, host-agnostic renderer so both TUIs share one implementation.

```go
// Styles carries the lipgloss styles a host wants applied to each element.
type Styles struct {
    H1, H2, H3, Bold, Italic, Code, Bullet lipgloss.Style
}

// Render returns ANSI-styled text. width <= 0 disables hard-wrapping
// (the host wraps instead).
func Render(src string, width int, st Styles) string
```

- Line-oriented: parse block elements, then inline spans, then word-wrap.
- `width <= 0` ⇒ no hard wrap. Bubble Tea passes the exact Detail inner width;
  tview passes `0` and lets its `TextView` wrap (sidesteps tview's pre-draw
  width problem).

**Supported subset (v1):**

- Headers: lines beginning `# `, `## `, `### ` (marker stripped, text styled).
- Bold: `**text**`.
- Italic: `*text*` or `_text_`.
- Inline code: `` `text` ``.
- Bullet lists: lines beginning `- ` or `* ` → rendered with a `•` bullet.
- Numbered lists: lines beginning `1. ` (etc.) → number preserved, indented.
- Blank line → paragraph break.
- Any other line → paragraph text, word-wrapped.

**Explicitly out of scope (YAGNI):** tables, blockquotes, fenced code blocks,
nested lists, link/image rendering (URLs shown as-is).

### 2. Detail pane wiring

- **bubbletui** (`view.go`): replace the raw `t.Notes` write with
  `markdown.Render(t.Notes, detailWidth, mdStyles)`, where `mdStyles` is built
  from the active theme.
- **tview** (`app.go`): `a.detail.SetText(tview.TranslateANSI(markdown.Render(
  t.Notes, 0, mdStyles)))` and set `SetDynamicColors(true)` on the Detail
  `TextView`. `TranslateANSI` converts the renderer's ANSI output into tview's
  style tags.

### 3. Multi-line input — bubbletui

The add/edit form currently stores fields as a homogeneous
`[]textinput.Model`. To mix a `textarea` for Notes without special-casing it
everywhere, introduce a small `formField` interface:

```go
type formField interface {
    Focus() tea.Cmd
    Blur()
    Update(tea.Msg) (formField, tea.Cmd)
    View() string
    Value() string
    SetValue(string)
    multiline() bool
}
```

Thin wrappers adapt `textinput.Model` (single-line) and `textarea.Model`
(multi-line). `m.inputs` becomes `[]formField`; focus/navigation/view code
stays generic.

**Submit-key conflict:** a textarea needs Enter for newlines. Rule:

- **Enter** submits the form **only when the focused field is single-line**.
- **Ctrl+S** always submits, from any field.
- Tab / Shift-Tab / Esc behave as today (already intercepted before the field
  receives the key).

Hint bars updated to advertise Ctrl+S and the Enter-inserts-newline behavior.

### 4. Multi-line input — tview

Nearly drop-in: replace the Notes `AddInputField` with
`AddTextArea("Notes", initial, 40, 6, 0, onChange)` in the add and edit forms.
Submission is already via the Add/Save **button**, so there is no Enter
conflict. The form hint is updated (Enter is no longer "select" while the
textarea is focused).

**Verification item:** confirm Tab still moves between form fields while the
textarea is focused in tview's pinned version. Fallback if not: a form-level
`SetInputCapture` that remaps Tab/Shift-Tab to field navigation.

## Testing

- `internal/markdown`: table-driven tests for each block/inline element, word
  wrapping at a given width, and the plain-text passthrough case.
- bubbletui: extend form tests for the `formField` abstraction, Ctrl+S submit,
  Enter-inserts-newline inside Notes, and Tab navigation across mixed field
  types.
- tview: extend form tests to confirm the Notes field accepts multi-line text
  and round-trips through `EditTask`.

## Non-goals

- No CLI changes.
- No markdown rendering in the task-list pane (titles stay plain).
- No fenced code blocks / tables / link rendering in v1.
