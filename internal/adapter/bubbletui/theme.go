package bubbletui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Theme is the color palette every style derives from. bg/fg are optional: when
// set, the theme paints a "page" background and ink foreground across the panes
// (a full-notebook look); when nil, panes render on the terminal's own colors.
type Theme struct {
	accent    lipgloss.TerminalColor // focused border, tags, titles
	secondary lipgloss.TerminalColor // unfocused borders, help keys
	subtle    lipgloss.TerminalColor // dim/help text
	selBg     lipgloss.TerminalColor // selected-row background
	selFg     lipgloss.TerminalColor // selected-row text
	warn      lipgloss.TerminalColor // status/errors
	bg        lipgloss.TerminalColor // optional page background (nil = terminal default)
	fg        lipgloss.TerminalColor // optional body ink (nil = terminal default)
}

// themeDefault is the original Charm look, adaptive to light/dark terminals.
var themeDefault = Theme{
	accent:    lipgloss.AdaptiveColor{Light: "205", Dark: "212"},
	secondary: lipgloss.AdaptiveColor{Light: "63", Dark: "99"},
	subtle:    lipgloss.AdaptiveColor{Light: "244", Dark: "245"},
	selBg:     lipgloss.AdaptiveColor{Light: "189", Dark: "57"},
	selFg:     lipgloss.AdaptiveColor{Light: "236", Dark: "231"},
	warn:      lipgloss.AdaptiveColor{Light: "160", Dark: "203"},
}

// themeNord — cool arctic blues/grays (fixed).
var themeNord = Theme{
	accent:    lipgloss.Color("110"),
	secondary: lipgloss.Color("109"),
	subtle:    lipgloss.Color("102"),
	selBg:     lipgloss.Color("24"),
	selFg:     lipgloss.Color("189"),
	warn:      lipgloss.Color("167"),
}

// themeDracula — dark with vivid purple/pink (fixed).
var themeDracula = Theme{
	accent:    lipgloss.Color("212"),
	secondary: lipgloss.Color("141"),
	subtle:    lipgloss.Color("103"),
	selBg:     lipgloss.Color("60"),
	selFg:     lipgloss.Color("231"),
	warn:      lipgloss.Color("203"),
}

// themeGruvbox — warm retro earth tones (fixed).
var themeGruvbox = Theme{
	accent:    lipgloss.Color("208"),
	secondary: lipgloss.Color("214"),
	subtle:    lipgloss.Color("245"),
	selBg:     lipgloss.Color("237"),
	selFg:     lipgloss.Color("223"),
	warn:      lipgloss.Color("167"),
}

// themeMono — grayscale + a red warn, adaptive to light/dark.
var themeMono = Theme{
	accent:    lipgloss.AdaptiveColor{Light: "238", Dark: "252"},
	secondary: lipgloss.AdaptiveColor{Light: "244", Dark: "245"},
	subtle:    lipgloss.AdaptiveColor{Light: "248", Dark: "240"},
	selBg:     lipgloss.AdaptiveColor{Light: "252", Dark: "238"},
	selFg:     lipgloss.AdaptiveColor{Light: "232", Dark: "255"},
	warn:      lipgloss.AdaptiveColor{Light: "160", Dark: "203"},
}

// themeNotebookDark — legal-pad accents on the terminal's own (dark) background:
// gold titles/borders and a pale-yellow "paper" highlight on the selected row,
// with no page background (fixed). This is the notebook look before the
// full-paper themeNotebook.
var themeNotebookDark = Theme{
	accent:    lipgloss.Color("220"), // gold — focused border, tags, titles
	secondary: lipgloss.Color("178"), // muted amber-gold
	subtle:    lipgloss.Color("136"), // dim amber
	selBg:     lipgloss.Color("229"), // pale legal-pad yellow highlight
	selFg:     lipgloss.Color("58"),  // dark olive ink on the highlight
	warn:      lipgloss.Color("203"), // red-pink for contrast
	// bg/fg left unset: renders on the terminal's own colors.
}

// themeNotebook — legal-pad look: navy ink on a cream-paper background, with
// gold accents and a bright-gold "highlighter" on the selected row (fixed).
// Accents are mid-gold (not bright) so titles/tags stay readable on cream paper.
var themeNotebook = Theme{
	accent:    lipgloss.Color("178"), // mid-gold — borders, titles, tags
	secondary: lipgloss.Color("136"), // darker gold — unfocused borders, help keys
	subtle:    lipgloss.Color("94"),  // brown — dim text, ruled lines
	selBg:     lipgloss.Color("220"), // bright-gold highlighter on the selected row
	selFg:     lipgloss.Color("17"),  // navy ink
	warn:      lipgloss.Color("160"), // red — reads on cream
	bg:        lipgloss.Color("230"), // cream paper
	fg:        lipgloss.Color("17"),  // navy ink
}

var themes = map[string]Theme{
	"default":       themeDefault,
	"nord":          themeNord,
	"dracula":       themeDracula,
	"gruvbox":       themeGruvbox,
	"mono":          themeMono,
	"notebook":      themeNotebook,
	"notebook-dark": themeNotebookDark,
}

// resolveTheme maps a name to a Theme; "" -> default; unknown -> error.
func resolveTheme(name string) (Theme, error) {
	if name == "" {
		name = "default"
	}
	if t, ok := themes[name]; ok {
		return t, nil
	}
	return Theme{}, fmt.Errorf("invalid theme %q (want default, nord, dracula, gruvbox, mono, notebook, notebook-dark)", name)
}
