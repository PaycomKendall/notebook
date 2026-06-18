# Bubble Tea TUI adapter (design)

**Date:** 2026-06-18
**Status:** Approved, ready for planning

## 1. Summary

Add a second, fully-featured TUI front-end for `notebook` built with the
**Bubble Tea / Lipgloss / Bubbles** (Charm) stack, alongside the existing
`tview` adapter. Both drive the same `internal/todo.Service`, so behavior
matches; only the rendering/engine differs. The engine is selected at runtime;
**tview remains the default**.

This is additive: the tview adapter is untouched except for one `launchTUI`
callback signature change (and its callers/tests).

## 2. Goals / non-goals

**Goals**
- Full feature parity with the tview TUI: three-pane layout (Lists | Tasks |
  Detail), pane navigation, task add/edit/delete/toggle (with tags + notes),
  list create/rename/delete, modal forms with Esc-cancel, and a context-aware
  footer.
- Runtime engine selection; default unchanged (tview).
- Keep the hexagonal boundary: the new adapter imports only `internal/todo` +
  Charm libs; only `cmd/nb/main.go` imports adapters.

**Non-goals**
- Replacing or modifying tview's behavior.
- Detail-pane scrolling (notes are short; rendered as static text — a small,
  deliberate simplification vs tview's scrollable TextView).
- Mouse support, themes/config, or animation beyond Bubble Tea defaults.

## 3. Engine selection & wiring

- The `tui` Cobra command gains `-e/--engine <tview|bubble>`.
- Resolution order (in the cli adapter, mirroring `resolveList`):
  flag → `$NB_TUI` → `"tview"`. An unrecognized value returns an error:
  `invalid engine %q (want "tview" or "bubble")`.
- `NewRootCmd`'s callback changes from `launchTUI func() error` to
  **`launchTUI func(engine string) error`**. The `tui` command resolves the
  engine and calls `launchTUI(engine)`. (Bare `nb` still prints help and does
  not launch a TUI.)
- `cmd/nb/main.go` dispatches:
  ```go
  launchTUI := func(engine string) error {
      if engine == "bubble" {
          return bubbletui.Run(svc)
      }
      return tui.New(svc).Run()
  }
  ```

## 4. Package structure (`internal/adapter/bubbletui`)

```
internal/adapter/bubbletui/
  model.go    Model struct, New, Init, Run, data-load helpers
  update.go   Update(msg): key routing by mode + window sizing
  view.go     View(): Lipgloss rendering (panes, footer, forms, confirm)
  styles.go   Lipgloss style/palette definitions
```

Exposes `func Run(svc *todo.Service) error` (runs a `tea.Program` with the
alt-screen). Implemented with a pointer-receiver `*Model` so `Update` mutates
in place and returns itself (clean for tests).

## 5. Model

```go
type focus int   // focusLists=0, focusTasks=1, focusDetail=2
type mode int    // modeNormal, modeAddTask, modeEditTask, modeNewList, modeRenameList, modeConfirm

type Model struct {
    svc *todo.Service

    listNames []string
    listIdx   int
    current   *todo.List // tasks for the selected list (nil if none)
    taskIdx   int

    focus focus
    mode  mode

    inputs   []textinput.Model // active form fields
    formField int              // focused field within inputs

    confirmPrompt string
    confirmAction func() error  // run on confirm

    status string // transient message (e.g. an error), shown in the footer
    width, height int
}
```

- `New(svc)` loads list names + the first list's tasks.
- `Init()` returns `nil` (Bubble Tea sends the initial `WindowSizeMsg`).
- Data-load helpers `reloadLists()`, `reloadTasks()` call the Service and clamp
  the selection indices into range.

## 6. Update — key handling

`Update` switches on `msg` type. On `tea.WindowSizeMsg` it stores `width/height`.
On `tea.KeyMsg` it routes by `mode`:

**modeNormal**
- `tab` / `shift+tab` → cycle `focus` (wraparound across the 3 panes).
- `up`/`k`, `down`/`j` → move the index within the focused pane (Lists →
  `listIdx` + reload tasks; Tasks → `taskIdx`), clamped.
- Tasks pane: `a` → modeAddTask, `d` → `svc.ToggleTask` + reload, `e`/`n` →
  modeEditTask (prefilled), `x` → modeConfirm (delete task).
- Lists pane: `a` → modeNewList, `r` → modeRenameList (prefilled), `x` →
  modeConfirm (delete list).
- `q` / `ctrl+c` → `tea.Quit`.

**form modes (add/edit/new-list/rename)**
- `tab`/`down` and `shift+tab`/`up` move `formField` across `inputs` (with
  focus blink); other keys go to the focused `textinput`.
- `enter` → validate, then perform the Service call for the current `mode`:
  - **Validation failure** (a required field — task title / list name — is
    empty after trim): the form **stays open**, no Service call.
  - **Service error** (e.g. duplicate list): the form **closes**, returning to
    modeNormal, and `status` is set to the error message.
  - **Success:** close the form, reload affected panes, clear `status`.
- `esc` → cancel, return to modeNormal.

**modeConfirm**
- `y`/`enter` → run `confirmAction`, reload, return to modeNormal.
- `n`/`esc` → cancel.

Service call mapping (identical to the tview adapter): `AddTask`,
`ToggleTask`, `EditTask`, `RemoveTask`, `CreateList`, `RenameList`,
`DeleteList`. Tags entered as a space-separated field → `strings.Fields`.

**Error handling:** any Service error (from a form submit or a direct action
like toggle/delete) sets `m.status` (shown in the footer in a warning style)
and the app never panics. A form closes on success or on a Service error;
it stays open only on a local validation failure (empty required field).

## 7. View — rendering

**modeNormal:** three Lipgloss-bordered panes joined with
`lipgloss.JoinHorizontal`:
- Lists: each name + open-count; selected row highlighted.
- Tasks: `[ ]`/`[x] #id title  #tags`; selected row highlighted; title for the
  current list.
- Detail: selected task's title, tags (accent), and notes (static text).
- The focused pane's border uses the accent color.
Below the panes, a context-aware footer (hints for the focused pane), or the
`status` message when one is set.

**form/confirm modes:** render the modal (rounded border, fields/buttons or
prompt) with its own hint line (`tab/↑↓ move · enter select · esc cancel`).
Pane widths are computed from `width` (Lists fixed ~18; Tasks/Detail split the
rest); sensible defaults before the first `WindowSizeMsg`.

## 8. Styling (`styles.go`)

Lipgloss styles: rounded borders; focused-pane border in an accent color
(pink), others in purple; selected-row highlight (light-on-purple); tags in
accent; footer keys bold/accent with dimmed labels; status in a warning color.
Matches the previewed render.

## 9. Dependencies

- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/lipgloss`
- `github.com/charmbracelet/bubbles` (textinput)

These raise no Go-version floor beyond the current 1.24.

## 10. Testing

Bubble Tea's `Update` is pure-ish and `View` returns a string, so the adapter
is unit-testable without a terminal:

- **update_test.go** — drive `Update` with `tea.KeyMsg` values and assert:
  focus cycles on Tab (wraps); `↑/↓`/`j`/`k` move and clamp indices; `d`
  toggles via the Service and persists; `a`→add form→`enter` adds a task;
  `e`→edit form→`enter` edits; `x`→confirm→`y` deletes; list `a`/`r`/`x` ops;
  `esc` cancels a form; `q` returns `tea.Quit`.
- **view_test.go** — assert `View()` contains list names, task lines with
  checkboxes, the selection marker, the footer hints, and (in form mode) the
  field labels + hint line.
- **model_test.go** — `New` loads lists/tasks; index clamping after reload.
- **cli_engine_test.go** — `tui --engine bubble`/`tview`, `$NB_TUI`, default
  `tview`, and an invalid value returns the expected error. (Tests inject a
  spy `launchTUI(engine string)` and assert the engine string received.)

The tview adapter and its tests are unchanged except for the `launchTUI`
signature update propagating to `cmd/nb` and the cli tests.
