# Remove tview — Bubble Tea-only TUI — design

**Date:** 2026-07-17
**Components:** `internal/adapter/tui`, `internal/adapter/cli`, `cmd/nb`, `demo.tape`, `README.md`, `go.mod`

## Summary

The project ships two interchangeable TUI engines: the tview adapter
(`internal/adapter/tui`) and the Bubble Tea adapter (`internal/adapter/bubbletui`),
selected at runtime by a `--engine` flag / `$NB_TUI` env var, defaulting to
tview. Bubble Tea is the actively-developed engine (themes, help icon, the
gopher easter egg) and the one in the README demo. This change removes the
tview engine entirely and makes Bubble Tea the only TUI. The engine-selection
machinery is deleted, not kept as a compatibility shim.

## Behavior change

- `nb tui` now always launches the Bubble Tea TUI (previously defaulted to
  tview).
- The `--engine`/`-e` flag and the `$NB_TUI` environment variable are removed.
  Existing invocations such as `nb tui --engine bubble` now fail with cobra's
  "unknown flag: --engine" error. This is accepted.
- The `--theme` flag and `$NB_THEME` are unchanged (themes were always a
  Bubble Tea feature); only the flag's help text is reworded to drop the
  now-redundant "bubble" qualifier.
- All non-TUI CLI commands (`add`, `ls`, `done`, etc.) are unaffected.

## Changes

### Delete
- `internal/adapter/tui/` — the entire tview adapter and its tests (`app.go`,
  `forms.go`, and the six `*_test.go` files).
- `internal/adapter/cli/cli_engine_test.go` — it exclusively tests
  `resolveEngine` and engine flag/env resolution, all of which are removed.

### `internal/adapter/cli/cli.go`
- Remove the `resolveEngine` function.
- Remove the `engine` variable, the `--engine`/`-e` flag registration, and any
  `$NB_TUI` lookup.
- Change `NewRootCmd`'s callback parameter from
  `launchTUI func(engine, theme string) error` to
  `launchTUI func(theme string) error`.
- The `tui` subcommand's `RunE` resolves only the theme (its existing theme
  rule: flag → `$NB_THEME` → default) and calls `launchTUI(theme)`.
- Reword the `--theme` flag help to drop "bubble" (e.g.
  `theme: default, nord, dracula, gruvbox, mono, notebook, notebook-dark; or $NB_THEME`).
- Drop any imports left unused after the removal (`fmt`, `os` — only if the
  compiler flags them; `os` likely stays for `$NB_THEME`).

### `cmd/nb/main.go`
- Remove the `internal/adapter/tui` import.
- Replace the engine-branching `launchTUI` closure with
  `launchTUI := func(theme string) error { return bubbletui.Run(svc, theme) }`.

### Test stubs (callback arity)
Update the `NewRootCmd` launchTUI stubs from two args to one:
- `cli_help_test.go` — `func(string, string) error` → `func(string) error`.
- `cli_robust_test.go` — same.
- `cli_test.go` — same.
- `cli_theme_test.go` — `func(_, th string) error` → `func(th string) error`
  (still captures the theme arg for its assertion).

### `go.mod` / `go.sum`
- Run `go mod tidy` to drop `github.com/rivo/tview` and
  `github.com/gdamore/tcell/v2` (direct) plus any now-orphaned transitive
  dependencies (e.g. `gdamore/encoding`, `lucasb-eyer/go-colorful` if unused
  elsewhere — tidy decides).

### `demo.tape`
- Line 30: `nb tui --engine bubble --theme notebook-dark` →
  `nb tui --theme notebook-dark`. No other lines change.

### `docs/demo.gif`
- Regenerate with `vhs demo.tape` after building/installing the updated `nb`
  binary (so the recorded command runs against the new flag surface). VHS is
  available at `/opt/homebrew/bin/vhs`.

### `README.md`
- Lines 29-31 currently read:
  ```
  nb tui                   # launch the interactive TUI (tview by default)
  nb tui --engine bubble   # launch the Bubble Tea TUI (or NB_TUI=bubble)
  nb tui --engine bubble --theme nord   # themes: ... (or NB_THEME)
  ```
  Replace with two lines that no longer mention engines:
  ```
  nb tui                        # launch the interactive TUI
  nb tui --theme nord           # themes: default, nord, dracula, gruvbox, mono, notebook, notebook-dark (or NB_THEME)
  ```

## Parity

Both adapters render the same `todo.Service` three-pane view (Folders / Pages /
Detail) with add/edit/new-list/rename/move/delete. Bubble Tea is a superset in
practice (themes, help icon, easter egg). No tview-only behavior is being
ported because there is none the project relies on. `go build ./...` +
`go test ./...` staying green is the safety net.

## Verification

1. `go build ./...` succeeds; `go test ./...` all pass.
2. `go mod tidy` leaves no tview/tcell entries: `grep -E 'tview|tcell' go.mod`
   returns nothing.
3. Repo-wide grep for stragglers returns nothing in `*.go`, `README.md`, and
   `demo.tape`: `tview`, `tcell`, `NB_TUI`, `resolveEngine`, `--engine`,
   `\-\-engine`, `adapter/tui`.
4. `docs/demo.gif` regenerated and renders the `nb tui --theme notebook-dark`
   session.

## Non-goals

- No changes to the Bubble Tea adapter's behavior or appearance.
- No new default theme (the built-in default theme is unchanged; the demo keeps
  `notebook-dark` via its flag).
- No compatibility shim for `--engine`/`$NB_TUI`.
