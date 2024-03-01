package routes

import (
	"context"
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/web/pages/todo"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

type todoEnv struct {
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

func (env *todoEnv) index(rw http.ResponseWriter, req *http.Request) error {
	slog.Debug("Request", slog.Any("request", req.URL.Path))
	return env.renderTodoApp(rw, req, true)
}

func (env *todoEnv) create(rw http.ResponseWriter, req *http.Request) error {
	newTodo := req.PostFormValue("new-todo")
	slog.Debug("newTodo", slog.String("newTodo", newTodo))

	usr := contextUser(req)
	input := &models.CreateTodoInput{
		UserId:    usr.Id,
		PlainText: newTodo,
		RichText:  newTodo,
	}

	_, err := env.todos.CreateNewTodo(req.Context(), input)
	if err != nil {
		return err
	}

	return env.renderTodoApp(rw, req, false)
}

func (env *todoEnv) delete(rw http.ResponseWriter, req *http.Request) error {
	todoId := req.PathValue("todoId")
	slog.Debug("todoId", slog.String("todoId", todoId))

	tid, err := strconv.ParseInt(todoId, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting todoId string to int: %w", err)
	}

	// make sure the user owns the todo
	todo, err := env.todos.FetchTodo(req.Context(), tid)
	if err != nil {
		return err
	}

	usr := contextUser(req)
	if todo.UserId != usr.Id {
		return fmt.Errorf("user does not own todo") // not authorized
	}

	err = env.todos.DeleteTodoById(req.Context(), tid)
	if err != nil {
		return err
	}

	return env.renderTodoApp(rw, req, false)
}

func (env *todoEnv) toggleCompleted(rw http.ResponseWriter, req *http.Request) error {
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

	usr := contextUser(req)
	if todo.UserId != usr.Id {
		return fmt.Errorf("user does not own todo") // not authorized
	}

	err = env.todos.ToggleTodoCompleted(req.Context(), todo)
	if err != nil {
		return err
	}

	return env.renderTodoApp(rw, req, false)
}

func (env *todoEnv) toggleAllCompleted(rw http.ResponseWriter, req *http.Request) error {
	c := req.PathValue("count")
	count, err := strconv.ParseInt(c, 10, 64)
	if err != nil {
		return err
	}
	slog.Debug("count", slog.Int64("count", count))

	usr := contextUser(req)
	err = env.todos.ToggleAllCompleted(req.Context(), usr.Id, count > 0)
	if err != nil {
		return err
	}

	return env.renderTodoApp(rw, req, false)
}

func (env *todoEnv) deleteCompleted(rw http.ResponseWriter, req *http.Request) error {
	usr := contextUser(req)
	err := env.todos.DeleteAllCompleted(req.Context(), usr.Id)
	if err != nil {
		return err
	}

	return env.renderTodoApp(rw, req, false)
}

func (env *todoEnv) renderTodoApp(rw http.ResponseWriter, req *http.Request, isFullPage bool) error {
	usr := contextUser(req)
	filter := req.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	todos, count, err := env.todos.FetchTodos(req.Context(), usr.Id, filter)
	if err != nil {
		return err
	}

	if isFullPage {
		todo.TodoAppPage(usr, todos, count).Render(req.Context(), rw)
	} else {
		todo.TodoApp(usr, todos, count).Render(req.Context(), rw)
	}

	return nil
}

func todoRoutes(rtr *router.Router, todos *models.TodoModel) {
	env := &todoEnv{
		todos: todos,
	}

	rtr.Group(func(r *router.Router) {
		r.Use(mustBeLoggedInMiddleware)
		r.Get("/todo", env.index)
		r.Post("/todo/create", env.create)
		r.Delete("/todo/{todoId}", env.delete)
		r.Patch("/todo/toggle-completed/{todoId}", env.toggleCompleted)
		r.Post("/todo/toggle-all/{count}", env.toggleAllCompleted)
		r.Delete("/todo/delete-completed", env.deleteCompleted)
	})
}
