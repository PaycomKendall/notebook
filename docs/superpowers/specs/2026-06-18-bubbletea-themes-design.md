# Bubble Tea themes (design)

**Date:** 2026-06-18
**Status:** Approved, ready for planning

## 1. Summary

Add selectable color themes to the Bubble Tea TUI adapter (`internal/adapter/bubbletui`).
Five presets — **default, nord, dracula, gruvbox, mono** — chosen at runtime via
`nb tui --theme <name>` or `$NB_THEME`, defaulting to `default`. Themes apply only
to the bubble engine (tview keeps its own look).

To support this cleanly, the adapter's styling moves from package-global
`lipgloss` vars onto the `Model` (a `Styles` struct built from a `Theme`).

## 2. Goals / non-goals

**Goals**
- Five themes selectable at runtime; default unchanged from today's look.
- `default` and `mono` adapt to light/dark terminals (`lipgloss.AdaptiveColor`);
  `nord`/`dracula`/`gruvbox` use their canonical fixed palettes.
- Keep the hexagonal boundary: theme validation lives in `bubbletui`; `cli` does
  not import it.

**Non-goals**
- Theming the tview adapter (different styling system).
- User-editable theme files or per-color overrides.
- A `--theme` effect on the tview engine (the flag is accepted but ignored there).

## 3. Selection & wiring

- The `tui` command gains `--theme <name>` (long flag only, to avoid the `-t`/tags
  association on other commands).
- Resolution (in the cli, raw string only): flag → `$NB_THEME` → `"default"`.
- `launchTUI` callback signature: `func(engine, theme string) error`.
- `cmd/nb/main.go`: `engine == "bubble"` → `bubbletui.Run(svc, theme)`; else
  `tui.New(svc).Run()` (theme ignored).
- `bubbletui.Run(svc, themeName)` calls `resolveTheme(themeName)` and returns its
  error before starting the program. Invalid name →
  `invalid theme "%s" (want default, nord, dracula, gruvbox, mono)`.
- Because validation lives in `bubbletui`, `--theme nope` errors only under
  `--engine bubble`; under tview the flag is ignored. (Documented.)

## 4. Theme & Styles model

```go
// theme.go
type Theme struct {
    accent    lipgloss.TerminalColor // focused border, tags, titles
    secondary lipgloss.TerminalColor // unfocused borders, help keys
    subtle    lipgloss.TerminalColor // dim/help text
    selBg     lipgloss.TerminalColor // selected-row background
    selFg     lipgloss.TerminalColor // selected-row text
    warn      lipgloss.TerminalColor // status/errors
}

var themes = map[string]Theme{ "default": ..., "nord": ..., "dracula": ..., "gruvbox": ..., "mono": ... }

func resolveTheme(name string) (Theme, error) // "" -> default; unknown -> error
```

```go
// styles.go (rewritten)
type Styles struct {
    title, dim, key, sel, tag, warn      lipgloss.Style
    pane, paneFocused, modal             lipgloss.Style
}

func (t Theme) styles() Styles // builds all styles from the theme's 6 colors
```

- `lipgloss.TerminalColor` lets a field hold either `lipgloss.Color` (fixed) or
  `lipgloss.AdaptiveColor{Light, Dark}` (default/mono).
- The package-global style vars are removed.

## 5. Model integration

- `Model` gains a `styles Styles` field.
- `New(svc *todo.Service, theme Theme) *Model` sets `m.styles = theme.styles()`.
- `view.go` / `forms.go` reference `m.styles.title` (etc.) instead of the old
  globals. The `hint` free function becomes a method `(m *Model) hint(pairs [][2]string) string`
  using `m.styles.key`/`m.styles.dim`.
- Test helper `newTestModel` passes `themeDefault` to `New`.

## 6. Presets (approximate ANSI-256; tunable by eye later)

| field | default (adaptive L/D) | nord | dracula | gruvbox | mono (adaptive L/D) |
|---|---|---|---|---|---|
| accent | 205 / 212 | 110 | 212 | 208 | 238 / 252 |
| secondary | 63 / 99 | 109 | 141 | 214 | 244 / 245 |
| subtle | 244 / 245 | 102 | 103 | 245 | 248 / 240 |
| selBg | 189 / 57 | 24 | 60 | 237 | 252 / 238 |
| selFg | 236 / 231 | 189 | 231 | 223 | 232 / 255 |
| warn | 160 / 203 | 167 | 203 | 167 | 160 / 203 |

(`default`/`mono` cells show `Light / Dark`; the others are fixed.)

## 7. Testing

- `resolveTheme`: `""`→default; each name returns a Theme; unknown→error. Presets
  are distinct (e.g. `themes["nord"].accent != themes["dracula"].accent`).
- `theme.styles()` integration: `New(svc, themes["nord"]).styles.title.GetForeground()`
  equals `themes["nord"].accent` (lipgloss exposes `GetForeground()`), proving styles
  derive from the chosen theme.
- Existing bubbletui tests keep passing via the updated `newTestModel` helper.
- Wiring: the spy `launchTUI` captures `(engine, theme)`; assert flag/env/default/
  override for `--theme`. (`resolveTheme`'s error path is unit-tested directly;
  `Run` can't be exercised headlessly.)
- Color is stripped in headless tests, so the live look needs a human check.

## 8. Files

- New: `internal/adapter/bubbletui/theme.go`.
- Rewritten: `internal/adapter/bubbletui/styles.go` (Styles struct + builder; globals removed).
- Modified: `model.go` (styles field, `New` signature), `view.go` + `forms.go`
  (use `m.styles`, `m.hint`), the bubbletui test files (helper + new theme tests).
- Modified (wiring): `cli.go` (`--theme` flag, raw resolution, `launchTUI` signature),
  `cmd/nb/main.go` (pass theme to `bubbletui.Run`), the cli test harnesses + engine test.
- Modified: `README.md` (document `--theme`).
