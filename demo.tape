# demo.tape — generates the README preview GIF for the notebook TUI.
#
# Requires: brew install vhs ffmpeg   (and `go install ./cmd/nb` so `nb` is on PATH)
# Run:      vhs demo.tape             -> writes docs/demo-tui.gif
#
# Re-run whenever the UI changes to regenerate a fresh preview.

Output docs/demo-tui.gif

Require nb

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "GitHub Dark"

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
Type "nb tui --theme notebook-dark" Enter
Sleep 2s

# Hop to the Folders pane and switch to the "inbox" folder
Shift+Tab
Sleep 800ms
Down@500ms 1
Sleep 1200ms

# Into the Pages list; move down through the pages
Tab
Sleep 800ms
Down@450ms 2
Sleep 900ms

# Toggle a page done, then undo
Type "d"
Sleep 1100ms
Type "d"
Sleep 900ms

# Back up to the page that has a note
Up@450ms 2
Sleep 900ms

# Open it on the ruled notebook page in the Detail pane
Tab
Sleep 2500ms

Type "q"
Sleep 1s
