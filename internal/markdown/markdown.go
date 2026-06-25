// internal/markdown/markdown.go
package markdown

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles carries the lipgloss styles a host applies to each markdown element.
type Styles struct {
	H1, H2, H3, Bold, Italic, Code, Bullet lipgloss.Style
}

var (
	reCode    = regexp.MustCompile("`([^`]+)`")
	reBold    = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reItalicA = regexp.MustCompile(`\*([^*]+)\*`)
	reItalicU = regexp.MustCompile(`_([^_]+)_`)
	reNumber  = regexp.MustCompile(`^(\d+)\.\s+(.*)$`)
)

// Render converts a markdown subset to ANSI-styled text. width <= 0 disables
// hard-wrapping (the host is responsible for wrapping).
func Render(src string, width int, st Styles) string {
	lines := strings.Split(src, "\n")
	out := make([]string, 0, len(lines))
	for _, raw := range lines {
		styled := renderBlock(raw, st)
		if width > 0 {
			styled = lipgloss.NewStyle().Width(width).Render(styled)
		}
		out = append(out, styled)
	}
	return strings.Join(out, "\n")
}

// renderBlock styles a single source line as a header, list item, or paragraph,
// then applies inline styling to its text.
func renderBlock(line string, st Styles) string {
	switch {
	case strings.HasPrefix(line, "### "):
		return st.H3.Render(inline(strings.TrimPrefix(line, "### "), st))
	case strings.HasPrefix(line, "## "):
		return st.H2.Render(inline(strings.TrimPrefix(line, "## "), st))
	case strings.HasPrefix(line, "# "):
		return st.H1.Render(inline(strings.TrimPrefix(line, "# "), st))
	case strings.HasPrefix(line, "- "), strings.HasPrefix(line, "* "):
		return st.Bullet.Render("•") + " " + inline(line[2:], st)
	}
	if m := reNumber.FindStringSubmatch(line); m != nil {
		return m[1] + ". " + inline(m[2], st)
	}
	return inline(line, st)
}

// inline applies code, bold, then italic styling. Code is processed first so
// markers inside backticks are left alone; bold before italic so ** is consumed
// before single-* italic. Injected ANSI never contains *, _, or ` so later
// passes cannot be confused by it.
func inline(s string, st Styles) string {
	s = reCode.ReplaceAllStringFunc(s, func(m string) string {
		return st.Code.Render(reCode.FindStringSubmatch(m)[1])
	})
	s = reBold.ReplaceAllStringFunc(s, func(m string) string {
		return st.Bold.Render(reBold.FindStringSubmatch(m)[1])
	})
	s = reItalicA.ReplaceAllStringFunc(s, func(m string) string {
		return st.Italic.Render(reItalicA.FindStringSubmatch(m)[1])
	})
	s = reItalicU.ReplaceAllStringFunc(s, func(m string) string {
		return st.Italic.Render(reItalicU.FindStringSubmatch(m)[1])
	})
	return s
}
