package tui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/kendallowen/notebook/internal/adapter/jsonstore"
	"github.com/kendallowen/notebook/internal/todo"
)

func newTestApp(t *testing.T) (*App, *todo.Service) {
	t.Helper()
	store, err := jsonstore.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	svc := todo.NewService(store)
	return New(svc), svc
}

func TestSkeletonRendersListsAndTasks(t *testing.T) {
	app, svc := newTestApp(t)
	if _, err := svc.AddTask("work", "ship it", nil, ""); err != nil {
		t.Fatal(err)
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(100, 24)

	app.buildUI()
	app.refreshLists()
	app.refreshTasks()
	app.refreshDetail()

	app.app.SetScreen(screen)
	go func() { _ = app.app.Run() }()
	defer app.app.Stop()
	app.app.Draw()

	if !screenContains(screen, "work") {
		t.Error("Lists pane should show 'work'")
	}
	if !screenContains(screen, "ship it") {
		t.Error("Tasks pane should show 'ship it'")
	}
}

// screenContains scans the simulation screen's cells for a substring on any row.
func screenContains(s tcell.SimulationScreen, want string) bool {
	cells, w, h := s.GetContents()
	for y := 0; y < h; y++ {
		row := make([]rune, 0, w)
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			if len(c.Runes) > 0 {
				row = append(row, c.Runes[0])
			} else {
				row = append(row, ' ')
			}
		}
		if containsRunes(string(row), want) {
			return true
		}
	}
	return false
}

func containsRunes(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
