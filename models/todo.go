package models

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type TodoModel struct {
	db      *sql.DB
	timeout time.Duration
}

func NewTodoModel(db *sql.DB, timeout time.Duration) *TodoModel {
	return &TodoModel{db: db, timeout: timeout}
}

type Todo struct {
	Id          int64
	UserId      int64
	PlainText   string
	RichText    string
	IsCompleted bool
}

type CreateTodoInput struct {
	UserId    int64
	PlainText string
	RichText  string
}

func (model *TodoModel) CreateNewTodo(c context.Context, input *CreateTodoInput) (*Todo, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	var lastInsertId int
	query := "INSERT INTO todos (user_id, plain_text, rich_text, is_completed) VALUES ($1, $2, $3, $4) RETURNING id"
	err := model.db.QueryRowContext(ctx, query, input.UserId, input.PlainText, input.RichText, false).Scan(&lastInsertId)
	if err != nil {
		return &Todo{}, err
	}

	newTodo := Todo{
		Id:          int64(lastInsertId),
		UserId:      input.UserId,
		PlainText:   input.PlainText,
		RichText:    input.RichText,
		IsCompleted: false,
	}

	slog.Info("newTodo", slog.Any("newTodo", newTodo))

	return &newTodo, nil
}

func (model *TodoModel) DeleteTodoById(c context.Context, todoId int64) error {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	result, err := model.db.ExecContext(ctx, "DELETE FROM todos WHERE id = ?", todoId)
	if err != nil {
		return fmt.Errorf("error deleting todo: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting number of rows affected: %w", err)
	}
	if rows != 1 {
		// warn, but continue execution
		slog.Warn("expected to affect 1 row, affected %d", rows)
	}

	return nil
}

func (model *TodoModel) FetchTodo(c context.Context, todoId int64) (*Todo, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	todo := Todo{}
	query := "SELECT id, user_id, plain_text, rich_text, is_completed FROM todos WHERE id = $1"
	err := model.db.QueryRowContext(ctx, query, todoId).Scan(&todo.Id, &todo.UserId, &todo.PlainText, &todo.RichText, &todo.IsCompleted)
	if err != nil {
		return &Todo{}, fmt.Errorf("error fetching todo: %w", err)
	}

	return &todo, nil
}

func (model *TodoModel) FetchTodos(c context.Context, userId int64, filter string) ([]Todo, int, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	var whereClause string
	switch filter {
	case "active":
		whereClause = "WHERE user_id = $1 AND is_completed = false"
	case "completed":
		whereClause = "WHERE user_id = $1 AND is_completed = true"
	default:
		whereClause = "WHERE user_id = $1"
	}

	qry := fmt.Sprintf("SELECT id, user_id, plain_text, rich_text, is_completed FROM todos %s", whereClause)

	rows, err := model.db.QueryContext(ctx, qry, userId)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	todos := make([]Todo, 0)

	for rows.Next() {
		todo := Todo{}
		if err := rows.Scan(&todo.Id, &todo.UserId, &todo.PlainText, &todo.RichText, &todo.IsCompleted); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			return nil, 0, err
		}
		todos = append(todos, todo)
	}
	// If the database is being written to, ensure to check for Close
	// errors that may be returned from the driver. The query may
	// encounter an auto-commit error and be forced to rollback changes.
	rerr := rows.Close()
	if rerr != nil {
		return nil, 0, rerr
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var activeCount int
	for i := range todos {
		if !todos[i].IsCompleted {
			activeCount++
		}
	}

	return todos, activeCount, nil
}

func (model *TodoModel) ToggleTodoCompleted(c context.Context, todo *Todo) error {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	result, err := model.db.ExecContext(ctx, "UPDATE todos SET is_completed = $1 WHERE id = $2", !todo.IsCompleted, todo.Id)
	if err != nil {
		return fmt.Errorf("error toggling todo completed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("[toggling todo completed] - error getting number of rows affected: %w", err)
	}
	if rows != 1 {
		// warn, but continue execution
		slog.Warn("[toggling todo completed] - expected to affect 1 row, affected %d", rows)
	}

	return nil
}

func (model *TodoModel) ToggleAllCompleted(c context.Context, userId int64, isCompleted bool) error {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	result, err := model.db.ExecContext(ctx, "UPDATE todos SET is_completed = $1 WHERE id = $2", isCompleted, userId)
	if err != nil {
		return fmt.Errorf("error toggling all todos: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("[toggling all todos] - error getting number of rows affected: %w", err)
	}
	if rows < 1 {
		// warn, but continue execution
		slog.Warn("[toggling all todos] - no rows affected")
	}

	return nil
}

func (model *TodoModel) DeleteAllCompleted(c context.Context, userId int64) error {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	result, err := model.db.ExecContext(ctx, "DELETE FROM todos WHERE user_id = $1 AND is_completed = true", userId)
	if err != nil {
		return fmt.Errorf("error deleting all completed todos: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("[deleting all completed todos] - error getting number of rows affected: %w", err)
	}
	if rows < 1 {
		// warn, but continue execution
		slog.Warn("[deleting all completed todos] - no rows affected")
	}

	return nil
}
