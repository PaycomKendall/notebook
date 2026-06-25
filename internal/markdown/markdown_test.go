// internal/markdown/markdown_test.go
package markdown

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// testStyles uses simple, detectable styling so assertions can look for the
// content text without depending on exact ANSI codes.
func testStyles() Styles {
	bold := lipgloss.NewStyle().Bold(true)
	return Styles{
		H1: bold, H2: bold, H3: bold,
		Bold: bold, Italic: lipgloss.NewStyle().Italic(true),
		Code:   lipgloss.NewStyle().Reverse(true),
		Bullet: lipgloss.NewStyle(),
	}
}

func TestRenderPlainPassthrough(t *testing.T) {
	out := Render("just text", 0, testStyles())
	if !strings.Contains(out, "just text") {
		t.Fatalf("plain text dropped: %q", out)
	}
}

func TestRenderHeaderStripsMarker(t *testing.T) {
	out := Render("# Title", 0, testStyles())
	if strings.Contains(out, "#") {
		t.Errorf("header marker not stripped: %q", out)
	}
	if !strings.Contains(out, "Title") {
		t.Errorf("header text missing: %q", out)
	}
}

func TestRenderBullet(t *testing.T) {
	out := Render("- milk", 0, testStyles())
	if !strings.Contains(out, "•") {
		t.Errorf("bullet glyph missing: %q", out)
	}
	if !strings.Contains(out, "milk") {
		t.Errorf("bullet text missing: %q", out)
	}
}

func TestRenderNumbered(t *testing.T) {
	out := Render("1. first", 0, testStyles())
	if !strings.Contains(out, "1.") || !strings.Contains(out, "first") {
		t.Errorf("numbered item malformed: %q", out)
	}
}

func TestRenderInlineMarkersConsumed(t *testing.T) {
	out := Render("a **b** c `d` _e_", 0, testStyles())
	for _, marker := range []string{"**", "`", "_"} {
		if strings.Contains(out, marker) {
			t.Errorf("inline marker %q not consumed: %q", marker, out)
		}
	}
	for _, text := range []string{"a", "b", "c", "d", "e"} {
		if !strings.Contains(out, text) {
			t.Errorf("inline text %q missing: %q", text, out)
		}
	}
}

func TestRenderWrapsToWidth(t *testing.T) {
	out := Render("one two three four five six seven eight", 10, testStyles())
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > 10 {
			t.Errorf("line exceeds width 10: %q (w=%d)", line, lipgloss.Width(line))
		}
	}
	if !strings.Contains(out, "\n") {
		t.Errorf("expected wrapping to produce multiple lines: %q", out)
	}
}
