package routes

import (
	"context"
	"eltimn/todo-plus/app/server/middleware"
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/web/pages/todo"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

const userId int64 = 1

type TodoEnv struct {
	todos interface {
		CreateNewTodo(c context.Context, input *models.CreateTodoInput) (*models.Todo, error)
		DeleteTodoById(c context.Context, todoId int64) error
		FetchTodo(c context.Context, todoId int64) (*models.Todo, error)
		FetchTodos(c context.Context, userId int64, filter string) ([]models.Todo, int, error)
		ToggleTodoCompleted(c context.Context, todo *models.Todo) error
		ToggleAllCompleted(c context.Context, userId int64, isCompleted bool) error
		DeleteAllCompleted(c context.Context, userId int64) error
	}
}

func (env *TodoEnv) index(w http.ResponseWriter, req *http.Request) error {
	slog.Debug("Request", slog.Any("request", req.URL.Path))
	env.renderTodoApp(w, req, true)
	return nil
}

func (env *TodoEnv) create(w http.ResponseWriter, req *http.Request) error {
	newTodo := req.PostFormValue("new-todo")
	slog.Debug("newTodo", slog.String("newTodo", newTodo))

	input := &models.CreateTodoInput{
		UserId:    userId,
		PlainText: newTodo,
		RichText:  newTodo,
	}

	_, err := env.todos.CreateNewTodo(req.Context(), input)
	if err != nil {
		return err
	}

	env.renderTodoApp(w, req, false)
	return nil
}

func (env *TodoEnv) delete(w http.ResponseWriter, req *http.Request) error {
	todoId := req.PathValue("todoId")
	slog.Debug("todoId", slog.String("todoId", todoId))

	tid, err := strconv.ParseInt(todoId, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting todoId string to int: %w", err)
	}

	err = env.todos.DeleteTodoById(req.Context(), tid)
	if err != nil {
		return err
	}

	env.renderTodoApp(w, req, false)
	return nil
}

func (env *TodoEnv) toggleCompleted(w http.ResponseWriter, req *http.Request) error {
	todoId := req.PathValue("todoId")
	slog.Info("todoId", slog.String("todoId", todoId))

	tid, err := strconv.ParseInt(todoId, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting todoId string to int: %w", err)
	}

	todo, err := env.todos.FetchTodo(req.Context(), tid)
	if err != nil {
		return err
	}

	err = env.todos.ToggleTodoCompleted(req.Context(), todo)
	if err != nil {
		return err
	}

	env.renderTodoApp(w, req, false)
	return nil
}

func (env *TodoEnv) toggleAllCompleted(w http.ResponseWriter, req *http.Request) error {
	c := req.PathValue("count")
	count, err := strconv.ParseInt(c, 10, 64)
	if err != nil {
		return err
	}
	slog.Debug("count", slog.Int64("count", count))

	err = env.todos.ToggleAllCompleted(req.Context(), userId, count > 0)
	if err != nil {
		return err
	}

	env.renderTodoApp(w, req, false)
	return nil
}

func (env *TodoEnv) deleteCompleted(w http.ResponseWriter, req *http.Request) error {
	err := env.todos.DeleteAllCompleted(req.Context(), userId)
	if err != nil {
		return err
	}

	env.renderTodoApp(w, req, false)
	return nil
}

func (env *TodoEnv) renderTodoApp(w http.ResponseWriter, req *http.Request, isFullPage bool) error {
	filter := req.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	todos, count, err := env.todos.FetchTodos(req.Context(), userId, filter)
	if err != nil {
		return err
	}

	if isFullPage {
		todo.TodoAppPage(todos, count).Render(req.Context(), w)
	} else {
		todo.TodoApp(todos, count).Render(req.Context(), w)
	}

	return nil
}

func todoRoutes(rtr *router.Router, todos *models.TodoModel) {
	env := &TodoEnv{
		todos: todos,
	}

	rtr.Group(func(r *router.Router) {
		r.Use(middleware.Mid(3))
		r.Get("/todo", env.index)
		r.Post("/todo/create", env.create)
		r.Delete("/todo/{todoId}", env.delete)
		r.Patch("/todo/toggle-completed/{todoId}", env.toggleCompleted)
		r.Post("/todo/toggle-all/{count}", env.toggleAllCompleted)
		r.Delete("/todo/delete-completed", env.deleteCompleted)
	})
}
