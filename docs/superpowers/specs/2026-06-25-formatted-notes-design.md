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
3. **"Notebook page" look for the Detail pane** — in the Bubble Tea TUI, dress
   the Detail pane to read like a sheet of notebook paper (header band, spiral
   binding gutter, margin rule, faint ruled lines), and widen it so the page
   has room to breathe.

## Decisions

- **Renderer:** a custom subset renderer in a new shared package. Chosen over
  `glamour` because the Detail pane is a narrow column in a three-pane layout
  (width/wrap control matters) and because glamour's own color themes would
  clash with the existing `nord`/`dracula`/`gruvbox`/`mono`/`default` themes.
  Zero new dependencies; fully unit-testable.
- **Scope:** both TUIs (tview default + Bubble Tea), matching the established
  feature-parity pattern, for the **multi-line input** and **markdown
  rendering**. The **"notebook page" Detail styling and the widened pane split
  are Bubble Tea only** (the user's explicit scope; tview's `TextView` makes
  per-line ruling/gutters awkward). The CLI is unchanged: `-n` already accepts
  any string (newlines pass through), and rendering is a display-only concern.
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

- **bubbletui** (`view.go`): render markdown, then wrap it in the notebook-page
  decoration (see §5). The markdown is produced with
  `markdown.Render(t.Notes, innerWidth, mdStyles)`, where `innerWidth` is the
  Detail width minus the binding gutter and margin rule, and `mdStyles` is built
  from the active theme.
- **tview** (`app.go`): `a.detail.SetText(tview.TranslateANSI(markdown.Render(
  t.Notes, 0, mdStyles)))` and set `SetDynamicColors(true)` on the Detail
  `TextView`. `TranslateANSI` converts the renderer's ANSI output into tview's
  style tags. tview gets markdown rendering only — **no** notebook chrome.

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

### 5. Notebook-page Detail styling — bubbletui only

`renderDetail` decorates the markdown output so the pane reads like notebook
paper. This is a line-oriented decoration layer applied around `markdown.Render`
output; it does not change the renderer. Elements (the approved **option D**,
all combined):

- **Header band:** a dim, letter-spaced label line beneath the pane title
  (e.g. `N O T E B O O K`) followed by a separator rule.
- **Spiral binding gutter:** a left column showing `◦ ` per content line, in the
  theme's subtle color.
- **Margin rule:** a colored vertical line between the gutter and the text
  (theme accent / a warm color), rendered per line.
- **Ruled lines:** each content line padded to the inner width and given a faint
  underline, so blank space below text still shows rules like lined paper.

Implementation notes: built per visual line in `renderDetail` using the active
theme's lipgloss styles. The width passed to `markdown.Render` is reduced by the
gutter + margin width so wrapping accounts for the chrome. Ruled lines fill the
pane's full height (`paneHeight`) so the page looks lined even past the note's
end. Empty-note state still shows a lined, empty page.

### 6. Widened Detail pane — bubbletui only

In `paneWidths` (`view.go`), flip the split of the available width (after the
fixed Lists column) from today's Tasks 3/5 ÷ Detail 2/5 to **Tasks 2/5 ÷ Detail
3/5**:

```go
tasks  = avail * 2 / 5
detail = avail - tasks
```

The Tasks pane only shows one-line titles, so it stays comfortable while the
notebook page gets the dominant share. tview layout is unchanged.

## Testing

- `internal/markdown`: table-driven tests for each block/inline element, word
  wrapping at a given width, and the plain-text passthrough case.
- bubbletui: extend form tests for the `formField` abstraction, Ctrl+S submit,
  Enter-inserts-newline inside Notes, and Tab navigation across mixed field
  types. Add a `paneWidths` test asserting Detail ≥ Tasks after the split flip,
  and a `renderDetail` test asserting the notebook chrome (gutter, header band)
  is present and that an empty note still produces a lined page.
- tview: extend form tests to confirm the Notes field accepts multi-line text
  and round-trips through `EditTask`.

## Non-goals

- No CLI changes.
- No markdown rendering in the task-list pane (titles stay plain).
- No fenced code blocks / tables / link rendering in v1.
- No notebook-page styling or pane-width change in the tview TUI (Bubble Tea
  only, per the user's scope).
