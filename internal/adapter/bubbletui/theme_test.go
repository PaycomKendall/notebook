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
