package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	todos, err := LoadTodos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load %s: %v\n", dataFile, err)
		todos = []Todo{}
	}

	p := tea.NewProgram(newModel(todos), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Safety net: changes are auto-saved as they happen, but make sure the
	// latest state hits disk on a clean exit.
	if m, ok := final.(model); ok {
		if err := SaveTodos(m.todos); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save todos: %v\n", err)
		}
	}
}
