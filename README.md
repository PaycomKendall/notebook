# notebook (`nb`)

A task tracker with a full CLI and an interactive three-pane TUI, built in Go
using a hexagonal (ports & adapters) architecture.

## Install

    brew install go            # if not already installed
    go install ./cmd/nb        # puts `nb` on ~/go/bin (ensure it's on PATH)

## Storage

One JSON file per list under (in order): `$NB_DIR`, `$XDG_DATA_HOME/notebook`,
or `~/.local/share/notebook`. Point `NB_DIR` at a git repo to version your notes.

## CLI

    nb                       # show help (banner + commands)
    nb tui                   # launch the interactive TUI (tview by default)
    nb tui --engine bubble   # launch the Bubble Tea TUI (or NB_TUI=bubble)
    nb tui --engine bubble --theme nord   # themes: default, nord, dracula, gruvbox, mono (or NB_THEME)
    nb add "buy milk" -l groceries -t store -n "2%"
    nb ls [-l list | -a] [-t tag] [--done|--open]
    nb done 3 [-l list]
    nb undone 3 [-l list]
    nb rm 3 [-l list]
    nb edit 3 --title "new" -n "note" [-l list]
    nb tag 3 --add urgent --rm home [-l list]
    nb mv 3 archive [-l work]   # move task #3 to "archive" (alias: move; dest auto-created)
    nb lists
    nb lists new ideas
    nb lists rename ideas later
    nb lists rm later --force

Default list is `inbox`; override with `-l` or `$NB_LIST`.

## TUI keys

A hint bar at the bottom shows the keys for the focused pane. The TUI opens
focused on the Tasks pane.

Panes: Lists | Tasks | Detail. `Tab`/`Shift-Tab` switch panes; `↑`/`↓` or `j`/`k` move within a pane.
Tasks: `a` add, `d` toggle done, `e`/`n` edit, `m` move to another list, `x` delete.
Lists: `a` new, `r` rename, `x` delete. `q`/`Ctrl-C` quit.
In forms/dialogs: `Tab`/`↑`/`↓` move between fields, `Enter` activates a button, `Esc` cancels.

## Development

    go test ./...
    go vet ./...
