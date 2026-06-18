package bubbletui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kendallowen/notebook/internal/todo"
)

type focusPane int

const (
	focusLists focusPane = iota
	focusTasks
	focusDetail
)

type mode int

const (
	modeNormal mode = iota
	modeAddTask
	modeEditTask
	modeNewList
	modeRenameList
	modeConfirm
)

// Model is the Bubble Tea front-end over the shared Service.
type Model struct {
	svc *todo.Service

	listNames []string
	listIdx   int
	current   *todo.List
	taskIdx   int

	focus focusPane
	mode  mode

	inputs      []textinput.Model
	formField   int
	formList    string
	formTaskID  int
	formOldName string

	confirmPrompt string
	confirmAction func() error

	status        string
	width, height int
}

// New builds a Model and loads the initial lists + tasks.
func New(svc *todo.Service) *Model {
	m := &Model{svc: svc, width: 90, height: 24}
	m.reloadLists()
	m.reloadTasks()
	return m
}

func (m *Model) reloadLists() {
	names, err := m.svc.ListNames()
	if err != nil {
		m.status = err.Error()
		return
	}
	m.listNames = names
	if m.listIdx >= len(names) {
		m.listIdx = len(names) - 1
	}
	if m.listIdx < 0 {
		m.listIdx = 0
	}
}

func (m *Model) currentListName() string {
	if m.listIdx < 0 || m.listIdx >= len(m.listNames) {
		return ""
	}
	return m.listNames[m.listIdx]
}

func (m *Model) reloadTasks() {
	name := m.currentListName()
	if name == "" {
		m.current = nil
		m.taskIdx = 0
		return
	}
	l, err := m.svc.GetList(name)
	if err != nil {
		m.current = nil
		m.taskIdx = 0
		return
	}
	m.current = l
	if len(l.Tasks) == 0 {
		m.taskIdx = 0
		return
	}
	if m.taskIdx >= len(l.Tasks) {
		m.taskIdx = len(l.Tasks) - 1
	}
	if m.taskIdx < 0 {
		m.taskIdx = 0
	}
}

func (m *Model) selectedTask() *todo.Task {
	if m.current == nil || m.taskIdx < 0 || m.taskIdx >= len(m.current.Tasks) {
		return nil
	}
	return &m.current.Tasks[m.taskIdx]
}

// Init satisfies tea.Model; Bubble Tea sends the initial WindowSizeMsg.
func (m *Model) Init() tea.Cmd { return nil }

// Update is a stub until Task 4.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

// Run starts the Bubble Tea program in the alternate screen.
func Run(svc *todo.Service) error {
	p := tea.NewProgram(New(svc), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
