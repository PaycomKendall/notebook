# Gopher easter egg — design

**Date:** 2026-07-17
**Component:** `internal/adapter/bubbletui`

## Summary

Add a hidden easter egg to the Bubble Tea TUI: when a user creates a page (task)
whose title is exactly `gopher`, the app briefly takes over the whole terminal
to display a gopher photo, rendered as truecolor half-block "pixels." Any key
dismisses it and returns the user to exactly where they were. The page is still
created normally — the takeover is purely visual.

## Trigger

- Fires when a **task is successfully added** with title exactly `gopher`.
- Match is **exact and case-sensitive** against the already-trimmed title:
  `strings.TrimSpace(title) == "gopher"`. So `gopher` triggers; `Gopher`,
  `GOPHER`, `gopher notes`, and `go pher` do not.
- Trigger location: `submitForm()` in `forms.go`, in the `modeAddTask` branch,
  after `AddTask` succeeds (no error) and after `closeForm()`. Because
  `closeForm()` resets `m.mode` to `modeNormal`, the mode switch to `modeGopher`
  happens *after* that reset.
- Applies in **any** folder — the folder name is irrelevant.
- Only the add-task path triggers it. Editing a task's title to `gopher`
  (`modeEditTask`) does **not** trigger the egg.

## Behavior

- On trigger, `m.mode` becomes a new `modeGopher`. The three panes are replaced
  by a full-terminal rendering of the gopher image, scaled to fit the current
  viewport and centered, with a dimmed hint line (e.g. `press any key to
  return`) at the bottom.
- **Any** key press in `modeGopher` returns to `modeNormal`. Focus, selected
  folder, and selected page are untouched — the user lands exactly where they
  were, with the new `gopher` page present in the list.
- `ctrl+c` still quits (handled before the any-key-returns logic, mirroring
  other modes).

## Rendering

Truecolor half-block rendering: each character cell is the `▀` (upper half
block) glyph with foreground = the upper pixel's color and background = the
lower pixel's color, giving two vertical pixels per cell in 24-bit color. This
displays the actual photo downsampled to the terminal's cell grid, is portable
to any truecolor terminal, and is plain styled text — so it needs no special
layout handling and drops into the normal `View()` flow.

- The source image is a **downscaled JPEG/PNG committed to the repo** and
  embedded with `//go:embed`, so the binary is self-contained (no runtime file
  dependency). Target embed size: small enough to keep the binary lean but large
  enough to look good when scaled down (a few hundred px on the long edge).
- Scaling: fit the image within the available `(width, height*2)` pixel budget
  (height doubled because each cell holds two vertical pixels) while preserving
  aspect ratio; center the result in the viewport.
- Color output uses lipgloss truecolor styles, consistent with the existing
  styling approach.

## Components (all in `internal/adapter/bubbletui/`)

- **`model.go`** — add `modeGopher` to the `mode` enum.
- **`gopher.go`** — new file:
  - `//go:embed` of the committed image asset.
  - A pure function, e.g. `renderGopher(width, height int) string`, that
    decodes the embedded image, downsamples to fit `(width, height)` cells, and
    returns the half-block string. Pure (no `Model` dependency) so it is
    unit-testable in isolation. Decode happens on demand; may be memoised later
    if needed (not required for v1).
  - Small helper `halfBlocks(img image.Image) string` (or equivalent) that does
    the per-cell `▀` conversion — the deterministic core to unit-test.
- **`docs`/asset** — the embedded image file lives alongside the package (e.g.
  `internal/adapter/bubbletui/gopher.jpg`) so `//go:embed` can reference it.
- **`view.go`** — add a `case modeGopher` to `View()` that returns the
  full-screen render (image + centered layout + hint line), sized to
  `m.width`/`m.height`.
- **`update.go`** — in `Update`'s key routing, add a `case modeGopher` that
  routes to a new `updateGopher(msg)` handler. `updateGopher` returns to
  `modeNormal` on any key (except `ctrl+c` → `tea.Quit`), mirroring
  `updateConfirm`.
- **`forms.go`** — in `submitForm()`'s `modeAddTask` branch, after the
  successful add + `closeForm()` + reloads, set `m.mode = modeGopher` when the
  trimmed title equals `gopher`.

## Testing

- **Unit — half-block converter:** feed a tiny known image (e.g. a 2×2 solid or
  checkerboard `image.RGBA`) to the converter and assert the exact output
  string (glyphs + expected color escapes). Deterministic.
- **Unit — fit/scale:** assert the rendered output respects the requested width
  and never exceeds the height budget for a given viewport.
- **Model test — trigger on add:** simulate adding a task titled `gopher` and
  assert `m.mode == modeGopher`; assert adding a task titled `Gopher` or
  `gopher notes` leaves `m.mode == modeNormal`.
- **Model test — dismiss:** from `modeGopher`, send an arbitrary key and assert
  `m.mode == modeNormal` and that focus/selection are unchanged.

## Non-goals

- No Kitty/iTerm2/sixel pixel protocols — half-block only for v1.
- No trigger on editing an existing task's title.
- No config/flag to enable/disable the egg; it is always on (it is hidden).
