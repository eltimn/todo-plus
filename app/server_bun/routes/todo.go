package routes

import (
	"eltimn/todo-plus/app/server_bun/models"
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

	err = models.CreateNewTodo(req.Request.Context(), uid, newTodo, newTodo)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func (th *todoHandler) delete(w http.ResponseWriter, req bunrouter.Request) error {
	params := req.Params()
	todoId := params.ByName("todoId")
	slog.Debug("todoId", slog.String("todoId", todoId))

	tid, err := primitive.ObjectIDFromHex(todoId)
	if err != nil {
		return err
	}

	err = models.DeleteTodoById(req.Request.Context(), tid)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func (th *todoHandler) toggleCompleted(w http.ResponseWriter, req bunrouter.Request) error {
	params := req.Params()
	todoId := params.ByName("todoId")
	slog.Info("todoId", slog.String("todoId", todoId))

	tid, err := primitive.ObjectIDFromHex(todoId)
	if err != nil {
		return err
	}

	todo, err := models.FetchTodo(req.Request.Context(), tid)
	if err != nil {
		return err
	}

	err = models.ToggleTodoCompleted(req.Request.Context(), todo)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func (th *todoHandler) toggleAllCompleted(w http.ResponseWriter, req bunrouter.Request) error {
	uid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}

	params := req.Params()
	count, err := params.Int64("count")
	if err != nil {
		return err
	}
	slog.Debug("count", slog.Int64("count", count))

	err = models.ToggleAllCompleted(req.Request.Context(), uid, count > 0)
	if err != nil {
		return err
	}

	renderTodoApp(w, req, false)
	return nil
}

func (th *todoHandler) deleteCompleted(w http.ResponseWriter, req bunrouter.Request) error {
	uid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}

	err = models.DeleteAllCompleted(req.Request.Context(), uid)
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

	todos, count, err := models.FetchTodos(req.Request.Context(), uid, "all")
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
	group.DELETE("/:todoId", todoHandler.delete)
	group.POST("/create", todoHandler.create)
	group.PATCH("/toggle-completed/:todoId", todoHandler.toggleCompleted)
	group.POST("/toggle-all/:count", todoHandler.toggleAllCompleted)
	group.DELETE("/delete-completed", todoHandler.deleteCompleted)
}
