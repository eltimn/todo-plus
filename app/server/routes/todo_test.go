package routes

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/router"
)

type mockTodoModel struct{}

func (m *mockTodoModel) CreateNewTodo(c context.Context, input *models.CreateTodoInput) (*models.Todo, error) {
	return nil, nil
}
func (m *mockTodoModel) DeleteTodoById(c context.Context, todoId int64) error { return nil }
func (m *mockTodoModel) FetchTodo(c context.Context, todoId int64) (*models.Todo, error) {
	return nil, nil
}

func (m *mockTodoModel) FetchTodos(c context.Context, userId int64, filter string) ([]models.Todo, int, error) {
	var todos []models.Todo

	todos = append(todos, models.Todo{1, 1, "message1", "message1", false})
	todos = append(todos, models.Todo{2, 1, "message2", "message2", false})

	return todos, 2, nil
}

func (m *mockTodoModel) ToggleTodoCompleted(c context.Context, todo *models.Todo) error { return nil }
func (m *mockTodoModel) ToggleAllCompleted(c context.Context, userId int64, isCompleted bool) error {
	return nil
}
func (m *mockTodoModel) DeleteAllCompleted(c context.Context, userId int64) error { return nil }

func TestTodosIndex(t *testing.T) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/todo", nil)

	env := TodoEnv{todos: &mockTodoModel{}}

	router.HttpHandler(env.index).ServeHTTP(rec, req)

	// expected := "1, 1, message1, message2, false"
	fmt.Println(rec.Body.String())
	// if expected != rec.Body.String() {
	// 	t.Errorf("\n...expected = %v\n...obtained = %v", expected, rec.Body.String())
	// }
}
