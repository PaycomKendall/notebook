package bubbletui

import "testing"

func TestStylesDeriveFromTheme(t *testing.T) {
	m, _ := newTestModel(t, nil)
	if got := m.styles.title.GetForeground(); got != themeDefault.accent {
		t.Errorf("title foreground = %v, want theme accent %v", got, themeDefault.accent)
	}
	if got := m.styles.sel.GetBackground(); got != themeDefault.selBg {
		t.Errorf("sel background = %v, want theme selBg %v", got, themeDefault.selBg)
	}
	if got := m.styles.paneFocused.GetBorderTopForeground(); got != themeDefault.accent {
		t.Errorf("focused border = %v, want accent %v", got, themeDefault.accent)
	}
}

func TestResolveThemeDefaultsAndKnown(t *testing.T) {
	got, err := resolveTheme("")
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if got != themeDefault {
		t.Error("empty name should resolve to the default theme")
	}
	for _, name := range []string{"default", "nord", "dracula", "gruvbox", "mono", "notebook", "notebook-dark"} {
		if _, err := resolveTheme(name); err != nil {
			t.Errorf("resolveTheme(%q) = %v", name, err)
		}
	}
}

func TestResolveThemeInvalid(t *testing.T) {
	if _, err := resolveTheme("nope"); err == nil {
		t.Error("invalid theme name should return an error")
	}
}

func TestThemesAreDistinct(t *testing.T) {
	if themes["nord"].accent == themes["dracula"].accent {
		t.Error("nord and dracula accents should differ")
	}
	if themes["gruvbox"].accent == themes["mono"].accent {
		t.Error("gruvbox and mono accents should differ")
	}
}

func TestStylesFollowChosenTheme(t *testing.T) {
	m, _ := newTestModel(t, nil)
	m2 := New(m.svc, themes["nord"])
	if got := m2.styles.title.GetForeground(); got != themeNord.accent {
		t.Errorf("nord title foreground = %v, want %v", got, themeNord.accent)
	}
}
