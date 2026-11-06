package main

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(r ...rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: r}
}

func update(m model, msg tea.Msg) model {
	next, _ := m.Update(msg)
	return next.(model)
}

func TestAddToggleDeletePersist(t *testing.T) {
	dir := t.TempDir()
	dataFile = filepath.Join(dir, "todos.json")
	t.Cleanup(func() { dataFile = defaultDataFile })

	m := newModel([]Todo{})

	// add "Buy milk"
	m = update(m, key('a'))
	m = update(m, key([]rune("Buy milk")...))
	m = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.todos) != 1 || m.todos[0].Text != "Buy milk" {
		t.Fatalf("expected 1 todo 'Buy milk', got %+v", m.todos)
	}

	// add "Water plants"
	m = update(m, key('a'))
	m = update(m, key([]rune("Water plants")...))
	m = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(m.todos))
	}

	// toggle first done (cursor is on the second item, move up first)
	m = update(m, key('k'))
	m = update(m, key('x'))
	if !m.todos[0].Completed {
		t.Fatal("expected first todo completed")
	}
	if m.todos[1].Completed {
		t.Fatal("expected second todo still active")
	}

	// move down, delete second
	m = update(m, key('j'))
	m = update(m, key('d'))
	if len(m.todos) != 1 || m.todos[0].Text != "Buy milk" {
		t.Fatalf("delete failed: %+v", m.todos)
	}
	if m.cursor != 0 {
		t.Fatalf("cursor should clamp to 0, got %d", m.cursor)
	}

	// persisted to disk and round-trips
	data, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatalf("todos.json not written: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty todos.json")
	}
	loaded, err := LoadTodos()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Text != "Buy milk" || !loaded[0].Completed {
		t.Fatalf("round-trip mismatch: %+v", loaded)
	}

	// view renders without error
	if v := m.View(); v == "" {
		t.Fatal("empty view")
	}
}
