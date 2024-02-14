package models

import (
	"log/slog"
)

// pwd: 3ClyP69h9ax1zNE

const collectionName = "todos"

type Todo struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	PlainText   string `json:"plain_text"`
	RichText    string `json:"rich_text"`
	IsCompleted bool   `json:"is_completed"`
}

func FetchTodos(userId string) ([]Todo, int, error) {
	todos := []Todo{
		{ID: "1", UserID: userId, PlainText: "Buy groceries", RichText: "Buy groceries", IsCompleted: true},
		{ID: "2", UserID: userId, PlainText: "Call mom", RichText: "Call mom", IsCompleted: false},
		{ID: "3", UserID: userId, PlainText: "Write blog post", RichText: "Write blog post", IsCompleted: false},
	}

	var activeCount int
	for i := range todos {
		if todos[i].IsCompleted {
			activeCount++
		}
	}

	// fmt.Println("todos", todos)

	return todos, activeCount, nil
}

func CreateNewTodo(userId, plainText, richText string) error {
	newTodo := Todo{
		UserID:      userId,
		PlainText:   plainText,
		RichText:    richText,
		IsCompleted: false,
	}

	slog.Info("newTodo", slog.Any("newTodo", newTodo))

	return nil
}
