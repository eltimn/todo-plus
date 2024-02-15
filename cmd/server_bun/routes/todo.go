package routes

import (
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/web/pages/todo"
	"log/slog"
	"net/http"

	"github.com/uptrace/bunrouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const userId string = "65cdb6ae82d84bf66d904c2c"

type todoHandler struct{}

func (th *todoHandler) index(w http.ResponseWriter, req bunrouter.Request) error {
	slog.Debug("Request", slog.Any("request", req.URL.Path))
	renderTodoApp(w, req, true)
	return nil
}

func (th *todoHandler) create(w http.ResponseWriter, req bunrouter.Request) error {
	newTodo := req.PostFormValue("new-todo")
	slog.Debug("newTodo", slog.String("newTodo", newTodo))

	uid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}

	err = models.CreateNewTodo(uid, newTodo, newTodo)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func renderTodoApp(w http.ResponseWriter, req bunrouter.Request, isFullPage bool) error {
	uid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}

	todos, count, err := models.FetchTodos(uid)
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

func todoRoutes(group *bunrouter.Group) {
	todoHandler := &todoHandler{}

	group.GET("", todoHandler.index)
	group.POST("/create", todoHandler.create)
}
