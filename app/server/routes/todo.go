package routes

import (
	"eltimn/todo-plus/app/server/middleware"
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/errs"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/web/pages/todo"
	"log/slog"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userId primitive.ObjectID

func init() {
	var err error
	userId, err = primitive.ObjectIDFromHex("65cdb6ae82d84bf66d904c2c")
	if err != nil {
		slog.Error("Could not initialize userId", errs.ErrAttr(err))
	}
}

type todoHandler struct{}

func (th *todoHandler) index(w http.ResponseWriter, req *http.Request) error {
	slog.Debug("Request", slog.Any("request", req.URL.Path))
	renderTodoApp(w, req, true)
	return nil
}

func (th *todoHandler) create(w http.ResponseWriter, req *http.Request) error {
	newTodo := req.PostFormValue("new-todo")
	slog.Debug("newTodo", slog.String("newTodo", newTodo))

	err := models.CreateNewTodo(req.Context(), userId, newTodo, newTodo)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func (th *todoHandler) delete(w http.ResponseWriter, req *http.Request) error {
	todoId := req.PathValue("todoId")
	slog.Debug("todoId", slog.String("todoId", todoId))

	tid, err := primitive.ObjectIDFromHex(todoId)
	if err != nil {
		return err
	}

	err = models.DeleteTodoById(req.Context(), tid)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func (th *todoHandler) toggleCompleted(w http.ResponseWriter, req *http.Request) error {
	todoId := req.PathValue("todoId")
	slog.Info("todoId", slog.String("todoId", todoId))

	tid, err := primitive.ObjectIDFromHex(todoId)
	if err != nil {
		return err
	}

	todo, err := models.FetchTodo(req.Context(), tid)
	if err != nil {
		return err
	}

	err = models.ToggleTodoCompleted(req.Context(), todo)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func (th *todoHandler) toggleAllCompleted(w http.ResponseWriter, req *http.Request) error {
	c := req.PathValue("count")
	count, err := strconv.ParseInt(c, 10, 64)
	if err != nil {
		return err
	}
	slog.Debug("count", slog.Int64("count", count))

	err = models.ToggleAllCompleted(req.Context(), userId, count > 0)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func (th *todoHandler) deleteCompleted(w http.ResponseWriter, req *http.Request) error {
	err := models.DeleteAllCompleted(req.Context(), userId)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func renderTodoApp(w http.ResponseWriter, req *http.Request, isFullPage bool) error {
	filter := req.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	todos, count, err := models.FetchTodos(req.Context(), userId, filter)
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

func todoRoutes(rtr *router.Router) {
	todoHandler := &todoHandler{}

	rtr.Group(func(r *router.Router) {
		r.Use(middleware.Mid(3))
		r.Get("/todo", todoHandler.index)
		r.Post("/todo/create", todoHandler.create)
		r.Delete("/todo/{todoId}", todoHandler.delete)
		r.Patch("/todo/toggle-completed/{todoId}", todoHandler.toggleCompleted)
		r.Post("/todo/toggle-all/{count}", todoHandler.toggleAllCompleted)
		r.Delete("/todo/delete-completed", todoHandler.deleteCompleted)
	})
}
