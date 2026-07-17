# Remove tview — Bubble Tea-only TUI — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the tview TUI engine entirely so Bubble Tea is the only interactive UI, and update docs + demo accordingly.

**Architecture:** Delete the `internal/adapter/tui` package and the engine-selection machinery in the CLI (`--engine`/`-e` flag, `$NB_TUI`, `resolveEngine`). `NewRootCmd`'s launch callback drops its `engine` parameter. `cmd/nb` wires the `tui` command straight to `bubbletui.Run`. Then prune dependencies and regenerate the README demo.

**Tech Stack:** Go 1.24, cobra CLI, Bubble Tea (`internal/adapter/bubbletui`), VHS (for the demo GIF, at `/opt/homebrew/bin/vhs`).

## Global Constraints

- Module path: `github.com/kendallowen/notebook`.
- Bubble Tea (`internal/adapter/bubbletui`) is the ONLY TUI after this change.
- No compatibility shim for `--engine`/`-e`/`$NB_TUI` — they are removed; old invocations may error.
- `--theme` flag and `$NB_THEME` are retained unchanged except help-text wording.
- This is a deletion/refactor: there is no red-first TDD cycle. The existing test suite is the regression net — the gate for each task is `go build ./...` and `go test ./...` passing.
- Verify commands: `go build ./...`, `go test ./...`.

---

### Task 1: Remove the tview engine and rewire the CLI to Bubble Tea only

This is one cohesive change: deleting the `tui` package breaks `main.go`'s import, and narrowing `NewRootCmd`'s callback breaks every stub simultaneously, so the tree only compiles once all edits are made together.

**Files:**
- Delete: `internal/adapter/tui/` (all 8 files: `app.go`, `forms.go`, `app_interact_test.go`, `app_test.go`, `footer_test.go`, `formkeys_test.go`, `forms_test.go`, `move_test.go`)
- Delete: `internal/adapter/cli/cli_engine_test.go`
- Modify: `internal/adapter/cli/cli.go`
- Modify: `cmd/nb/main.go`
- Modify: `internal/adapter/cli/cli_help_test.go`, `internal/adapter/cli/cli_robust_test.go`, `internal/adapter/cli/cli_test.go`, `internal/adapter/cli/cli_theme_test.go`

**Interfaces:**
- Produces: `NewRootCmd(svc *todo.Service, launchTUI func(theme string) error) *cobra.Command` — the callback loses its `engine` parameter.

- [ ] **Step 1: Delete the tview adapter package**

```bash
git rm -r internal/adapter/tui
```

- [ ] **Step 2: Delete the engine-resolution test**

```bash
git rm internal/adapter/cli/cli_engine_test.go
```

- [ ] **Step 3: Remove `resolveEngine` and the engine flag from `cli.go`**

In `internal/adapter/cli/cli.go`:

Delete the entire `resolveEngine` function (the block from the `// resolveEngine applies…` comment through its closing brace):

```go
// resolveEngine applies the engine rule: flag, then $NB_TUI, then "tview".
func resolveEngine(flag string) (string, error) {
	e := flag
	if e == "" {
		e = os.Getenv("NB_TUI")
	}
	if e == "" {
		e = "tview"
	}
	switch e {
	case "tview", "bubble":
		return e, nil
	default:
		return "", fmt.Errorf("invalid engine %q (want \"tview\" or \"bubble\")", e)
	}
}
```

Change the `NewRootCmd` signature and its doc comment from:

```go
// NewRootCmd builds the command tree. launchTUI runs the interactive UI for
// the chosen engine; bare `nb` prints help.
func NewRootCmd(svc *todo.Service, launchTUI func(engine, theme string) error) *cobra.Command {
```

to:

```go
// NewRootCmd builds the command tree. launchTUI runs the interactive UI;
// bare `nb` prints help.
func NewRootCmd(svc *todo.Service, launchTUI func(theme string) error) *cobra.Command {
```

Replace the `tui` command block. From:

```go
	var engine, theme string
	tui := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := resolveEngine(engine)
			if err != nil {
				return err
			}
			th := theme
			if th == "" {
				th = os.Getenv("NB_THEME")
			}
			if th == "" {
				th = "default"
			}
			return launchTUI(eng, th)
		},
	}
	tui.Flags().StringVarP(&engine, "engine", "e", "", `TUI engine: "tview" (default) or "bubble"; or $NB_TUI`)
	tui.Flags().StringVar(&theme, "theme", "", `bubble theme: default, nord, dracula, gruvbox, mono, notebook, notebook-dark; or $NB_THEME`)
```

to:

```go
	var theme string
	tui := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			th := theme
			if th == "" {
				th = os.Getenv("NB_THEME")
			}
			if th == "" {
				th = "default"
			}
			return launchTUI(th)
		},
	}
	tui.Flags().StringVar(&theme, "theme", "", `theme: default, nord, dracula, gruvbox, mono, notebook, notebook-dark; or $NB_THEME`)
```

Note: `fmt`, `os`, and `strings` imports all remain used (`buildBanner` uses `fmt`/`strings`; `resolveList` and the theme lookup use `os`), so the import block is unchanged.

- [ ] **Step 4: Rewire `cmd/nb/main.go` to Bubble Tea only**

Replace the whole file with:

```go
package main

import (
	"fmt"
	"os"

	"github.com/kendallowen/notebook/internal/adapter/bubbletui"
	"github.com/kendallowen/notebook/internal/adapter/cli"
	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

func main() {
	dir := os.Getenv("NB_DIR")
	store, err := jsonstore.New(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	svc := todo.NewService(store)

	launchTUI := func(theme string) error {
		return bubbletui.Run(svc, theme)
	}

	root := cli.NewRootCmd(svc, launchTUI)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5: Update the four test stubs to the one-arg callback**

- `internal/adapter/cli/cli_help_test.go`: change
  `NewRootCmd(svc, func(string, string) error { launched = true; return nil })`
  to
  `NewRootCmd(svc, func(string) error { launched = true; return nil })`
- `internal/adapter/cli/cli_robust_test.go`: change
  `NewRootCmd(svc, func(string, string) error { return nil })`
  to
  `NewRootCmd(svc, func(string) error { return nil })`
- `internal/adapter/cli/cli_test.go`: change
  `NewRootCmd(svc, func(string, string) error { return nil })`
  to
  `NewRootCmd(svc, func(string) error { return nil })`
- `internal/adapter/cli/cli_theme_test.go`: change
  `NewRootCmd(svc, func(_, th string) error { got = th; return nil })`
  to
  `NewRootCmd(svc, func(th string) error { got = th; return nil })`

- [ ] **Step 6: Build and run the full suite**

Run: `go build ./... && go test ./...`
Expected: build succeeds (no reference to `internal/adapter/tui`, no undefined `resolveEngine`); all packages PASS. `go vet ./...` clean.

If any test in `internal/adapter/cli` still references the removed engine behavior beyond the stubs above, fix it to match the new one-arg callback and rerun. Do not re-add engine logic.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: remove tview engine; Bubble Tea is the only TUI

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: Prune tview/tcell dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Tidy modules**

```bash
go mod tidy
```

- [ ] **Step 2: Verify tview/tcell are gone and the build still passes**

Run:
```bash
grep -nE 'tview|tcell' go.mod || echo "clean: no tview/tcell"
go build ./... && go test ./...
```
Expected: grep prints `clean: no tview/tcell`; build + tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: drop tview/tcell dependencies after engine removal

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: Update README and regenerate the demo

**Files:**
- Modify: `README.md` (lines 29-31)
- Modify: `demo.tape` (line 30)
- Regenerate: `docs/demo.gif`

- [ ] **Step 1: Update the README CLI examples**

In `README.md`, replace these three lines:

```
    nb tui                   # launch the interactive TUI (tview by default)
    nb tui --engine bubble   # launch the Bubble Tea TUI (or NB_TUI=bubble)
    nb tui --engine bubble --theme nord   # themes: default, nord, dracula, gruvbox, mono, notebook, notebook-dark (or NB_THEME)
```

with these two:

```
    nb tui                        # launch the interactive TUI
    nb tui --theme nord           # themes: default, nord, dracula, gruvbox, mono, notebook, notebook-dark (or NB_THEME)
```

- [ ] **Step 2: Update the demo tape command**

In `demo.tape`, change line 30 from:

```
Type "nb tui --engine bubble --theme notebook-dark" Enter
```

to:

```
Type "nb tui --theme notebook-dark" Enter
```

- [ ] **Step 3: Install the updated binary and regenerate the GIF**

The tape's `Require nb` runs the `nb` on `PATH`, so install the freshly-built binary first:

```bash
go install ./cmd/nb
vhs demo.tape
```
Expected: `vhs` writes `docs/demo.gif` with no errors; the recorded terminal shows `nb tui --theme notebook-dark` launching the Bubble Tea TUI.

- [ ] **Step 4: Sanity-check the artifact**

Run: `ls -la docs/demo.gif && git status --short`
Expected: `docs/demo.gif` is modified (regenerated); `README.md` and `demo.tape` show as modified.

- [ ] **Step 5: Commit**

```bash
git add README.md demo.tape docs/demo.gif
git commit -m "docs: drop engine flag from README/demo; regenerate GIF via nb tui

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: Whole-repo verification

**Files:** none (verification only).

- [ ] **Step 1: Confirm no stragglers remain**

Run:
```bash
grep -rnE 'tview|tcell|NB_TUI|resolveEngine|--engine|adapter/tui' \
  --include='*.go' --include='*.md' --include='*.tape' . \
  | grep -v 'docs/superpowers/' || echo "clean: no stragglers"
```
Expected: `clean: no stragglers` (the `docs/superpowers/` spec/plan legitimately mention these words and are excluded).

- [ ] **Step 2: Final build + suite**

Run: `go build ./... && go test ./...`
Expected: build succeeds; all packages PASS.

---

## Self-Review

**Spec coverage:**
- Delete tview adapter → Task 1 Step 1. ✓
- Delete cli_engine_test.go → Task 1 Step 2. ✓
- cli.go: remove resolveEngine/engine flag/$NB_TUI, narrow callback, reword theme help → Task 1 Step 3. ✓
- main.go rewire → Task 1 Step 4. ✓
- Four test stubs → Task 1 Step 5. ✓
- go mod tidy dropping tview/tcell → Task 2. ✓
- README lines 29-31 → Task 3 Step 1. ✓
- demo.tape line 30 → Task 3 Step 2. ✓
- Regenerate docs/demo.gif via vhs → Task 3 Step 3. ✓
- Verification (build/test, grep clean, go.mod clean) → Task 2 Step 2, Task 4. ✓
- Behavior change (`nb tui` = Bubble Tea; `--engine` errors) → realized by Task 1; no shim added. ✓
- Non-goals (no bubble behavior change, no new default theme, no shim) → respected. ✓

**Placeholder scan:** No TBD/TODO; every code/edit step shows exact before/after content and exact commands. ✓

**Type consistency:** `NewRootCmd(..., launchTUI func(theme string) error)` is defined in Task 1 Step 3 and consumed identically in main.go (Step 4) and all four test stubs (Step 5). ✓
