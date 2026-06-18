package bubbletui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Theme is the six-color palette every style derives from.
type Theme struct {
	accent    lipgloss.TerminalColor // focused border, tags, titles
	secondary lipgloss.TerminalColor // unfocused borders, help keys
	subtle    lipgloss.TerminalColor // dim/help text
	selBg     lipgloss.TerminalColor // selected-row background
	selFg     lipgloss.TerminalColor // selected-row text
	warn      lipgloss.TerminalColor // status/errors
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

var themes = map[string]Theme{
	"default": themeDefault,
	"nord":    themeNord,
	"dracula": themeDracula,
	"gruvbox": themeGruvbox,
	"mono":    themeMono,
}

// resolveTheme maps a name to a Theme; "" -> default; unknown -> error.
func resolveTheme(name string) (Theme, error) {
	if name == "" {
		name = "default"
	}
	if t, ok := themes[name]; ok {
		return t, nil
	}
	return Theme{}, fmt.Errorf("invalid theme %q (want default, nord, dracula, gruvbox, mono)", name)
}
