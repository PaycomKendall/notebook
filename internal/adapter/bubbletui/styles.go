package bubbletui

import "github.com/charmbracelet/lipgloss"

var (
	accent = lipgloss.Color("212")
	mauve  = lipgloss.Color("99")
	subtle = lipgloss.Color("245")
	selBg  = lipgloss.Color("57")
	white  = lipgloss.Color("231")
	warnFg = lipgloss.Color("203")

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(accent)
	dimStyle   = lipgloss.NewStyle().Foreground(subtle)
	keyStyle   = lipgloss.NewStyle().Bold(true).Foreground(mauve)
	selStyle   = lipgloss.NewStyle().Foreground(white).Background(selBg).Bold(true)
	tagStyle   = lipgloss.NewStyle().Foreground(accent)
	warnStyle  = lipgloss.NewStyle().Foreground(warnFg)

	paneStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(mauve).Padding(0, 1)
	paneFocused = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(0, 1)
	modalStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 2)
)
