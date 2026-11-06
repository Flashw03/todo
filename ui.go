package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().Padding(0, 1)

	doneStyle = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)

	selectedDoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Padding(0, 1)

	emptyStyle = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Padding(0, 1)
)

// clearMsg dismisses the status message after it has been shown.
type clearMsg struct{}

// model holds the UI state.
type model struct {
	todos   []Todo
	cursor  int
	adding  bool
	input   textinput.Model
	width   int
	height  int
	message string
}

func newModel(todos []Todo) model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = "what needs to be done?"
	ti.CharLimit = 300
	return model{
		todos:  todos,
		input:  ti,
		width:  80,
		height: 24,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case clearMsg:
		m.message = ""
		return m, nil

	case tea.KeyMsg:
		if m.adding {
			switch msg.String() {
			case "enter":
				return m.commit()
			case "esc":
				m.adding = false
				m.input.Blur()
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		cmd := tea.Cmd(nil)

		// Character keys arrive as KeyRunes; KeyMsg.String() reports them as
		// "runes", so match on the rune itself.
		if msg.Type == tea.KeyRunes {
			if len(msg.Runes) == 0 {
				return m, nil
			}
			switch msg.Runes[0] {
			case 'q':
				return m, tea.Quit
			case 'j':
				m.move(1)
			case 'k':
				m.move(-1)
			case 'g':
				m.cursor = 0
			case 'G':
				if len(m.todos) > 0 {
					m.cursor = len(m.todos) - 1
				}
			case 'a':
				m.adding = true
				m.input.Focus()
				return m, textinput.Blink
			case 'x':
				if len(m.todos) > 0 {
					m.todos[m.cursor].Completed = !m.todos[m.cursor].Completed
					cmd = m.afterSave()
				}
			case 'd':
				if len(m.todos) > 0 {
					m.todos = append(m.todos[:m.cursor], m.todos[m.cursor+1:]...)
					if m.cursor >= len(m.todos) {
						m.cursor = max(0, len(m.todos)-1)
					}
					cmd = m.afterSave()
				}
			}
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "down":
			m.move(1)
		case "up":
			m.move(-1)
		case "home":
			m.cursor = 0
		case "end":
			if len(m.todos) > 0 {
				m.cursor = len(m.todos) - 1
			}
		case "space":
			if len(m.todos) > 0 {
				m.todos[m.cursor].Completed = !m.todos[m.cursor].Completed
				cmd = m.afterSave()
			}
		case "delete", "backspace":
			if len(m.todos) > 0 {
				m.todos = append(m.todos[:m.cursor], m.todos[m.cursor+1:]...)
				if m.cursor >= len(m.todos) {
					m.cursor = max(0, len(m.todos)-1)
				}
				cmd = m.afterSave()
			}
		}
		return m, cmd
	}
	return m, nil
}

// move shifts the cursor by delta, clamped to the list bounds.
func (m *model) move(delta int) {
	if m.cursor+delta < 0 || m.cursor+delta > len(m.todos)-1 {
		return
	}
	m.cursor += delta
}

// afterSave persists the todos and, on failure, shows a status message that
// clears itself after a few seconds.
func (m *model) afterSave() tea.Cmd {
	m.message = saveMsg(SaveTodos(m.todos))
	if m.message == "" {
		return nil
	}
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return clearMsg{} })
}

// commit stores the text currently in the input as a new todo.
func (m model) commit() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	m.adding = false
	m.input.Blur()
	m.input.SetValue("")
	if text != "" {
		m.todos = append(m.todos, Todo{
			ID:        nextID(m.todos),
			Text:      text,
			CreatedAt: time.Now(),
		})
		m.cursor = len(m.todos) - 1
		cmd := m.afterSave()
		return m, cmd
	}
	return m, nil
}

// saveMsg converts a save error into a status-line message.
func saveMsg(err error) string {
	if err != nil {
		return "save failed: " + err.Error()
	}
	return ""
}

func (m model) View() string {
	lines := []string{}

	lines = append(lines, titleStyle.Render(" Todo"))
	lines = append(lines, dividerStyle.Render(strings.Repeat("─", max(1, m.width-2))))
	lines = append(lines, "")

	if len(m.todos) == 0 {
		lines = append(lines, emptyStyle.Render("  no tasks — press a to add one"))
	} else {
		visible := max(1, m.height-6)
		start := m.cursor
		if start > len(m.todos)-visible {
			start = max(0, len(m.todos)-visible)
		}
		for i := start; i < min(start+visible, len(m.todos)); i++ {
			t := m.todos[i]
			marker, label := "☐", "  "
			if t.Completed {
				marker = "✔"
			}
			if i == m.cursor {
				label = "▶ "
			}
			var row string
			switch {
			case i == m.cursor && t.Completed:
				row = selectedDoneStyle.Render(label + marker + " " + t.Text)
			case i == m.cursor:
				row = selectedStyle.Render(label + marker + " " + t.Text)
			case t.Completed:
				row = doneStyle.Render(label + marker + " " + t.Text)
			default:
				row = normalStyle.Render(label + marker + " " + t.Text)
			}
			lines = append(lines, row)
		}
	}

	lines = append(lines, "")

	if m.message != "" {
		lines = append(lines, warnStyle.Render(m.message))
	}

	if m.adding {
		lines = append(lines, m.input.View())
	} else {
		done := 0
		for _, t := range m.todos {
			if t.Completed {
				done++
			}
		}
		footer := fmt.Sprintf("%d tasks, %d done", len(m.todos), done)
		if m.width > 50 {
			footer += "   a add · x done · d del · j/k nav · g/G top/bottom · q quit"
		}
		lines = append(lines, footerStyle.Render(footer))
	}

	lines = append(lines, dividerStyle.Render(strings.Repeat("─", max(1, m.width-2))))
	return strings.Join(lines, "\n")
}
