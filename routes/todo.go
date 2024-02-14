package routes

import (
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/web/pages/todo"
	"log/slog"
	"net/http"
)

const userId = "blv7133ov73uoau"

func renderTodoApp(w http.ResponseWriter, r *http.Request, isFullPage bool) *appError {
	todos, count, err := models.FetchTodos(userId)
	if err != nil {
		return &appError{Message: "Error fetching todos", Code: http.StatusInternalServerError, error: err}
	}

	if isFullPage {
		todo.TodoAppPage(todos, count).Render(r.Context(), w)
	} else {
		todo.TodoApp(todos, count).Render(r.Context(), w)
	}

	return nil
}

func todoHandler(w http.ResponseWriter, r *http.Request) *appError {
	slog.Debug("Request", slog.Any("request", r.URL.Path))
	if r.URL.Path == "/" {
		renderTodoApp(w, r, true)
	}
	return nil
}

func createTodoHandler(w http.ResponseWriter, r *http.Request) *appError {
	newTodo := r.PostFormValue("new-todo")
	slog.Debug("newTodo", slog.String("newTodo", newTodo))

	err := models.CreateNewTodo(userId, newTodo, newTodo)
	if err != nil {
		return &appError{Message: "Error creating new todo", Code: http.StatusInternalServerError, error: err}
	}

	return &appError{Message: "Error creating new todo", Code: http.StatusInternalServerError, error: err}

	// renderTodoApp(w, r, false)
	// return nil
}

func todoRoutes() {
	http.Handle("GET /", appHandler(todoHandler))
	http.Handle("POST /todo/create", appHandler(createTodoHandler))
}
