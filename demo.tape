# demo.tape — generates the README preview GIF for the notebook TUI.
#
# Requires: brew install vhs ffmpeg   (and `go install ./cmd/nb` so `nb` is on PATH)
# Run:      vhs demo.tape             -> writes docs/demo.gif
#
# Re-run whenever the UI changes to regenerate a fresh preview.

Output docs/demo.gif

Require nb

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "nord"

# --- Setup (hidden from the recording): isolated data dir + sample tasks ---
Hide
Type "export NB_DIR=$(mktemp -d)" Enter
Type "nb add 'Draft launch post' -t writing -n 'Cover the notebook rename'" Enter
Type "nb add 'Review PR #12' -t code" Enter
Type "nb add 'Book dentist'" Enter
Type "nb add 'Buy oat milk' -l groceries -t store" Enter
Type "clear" Enter
Show

# --- Demo ---
Type "nb tui --engine bubble --theme nord" Enter
Sleep 2s

# Move down through the task list
Down@500ms 2
Sleep 1s

# Toggle a task done, then back
Type "d"
Sleep 1200ms
Type "d"
Sleep 1s

# Switch to the Detail pane to show the note
Tab
Sleep 1500ms

# Back to Tasks and quit
Tab
Sleep 800ms
Type "q"
Sleep 1s
