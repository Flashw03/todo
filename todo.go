package main

import (
	"encoding/json"
	"os"
	"time"
)

const defaultDataFile = "todos.json"

// dataFile is where todos are persisted. Overridable in tests.
var dataFile = defaultDataFile

// Todo represents a single task.
type Todo struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// LoadTodos reads todos from the JSON file. A missing file yields an empty
// slice so the app can start fresh.
func LoadTodos() ([]Todo, error) {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Todo{}, nil
		}
		return nil, err
	}
	var todos []Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

// SaveTodos writes all todos to the JSON file.
func SaveTodos(todos []Todo) error {
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, data, 0o644)
}

// nextID returns an ID greater than any existing one.
func nextID(todos []Todo) int64 {
	var max int64
	for _, t := range todos {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}
