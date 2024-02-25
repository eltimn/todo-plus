package models

import (
	"context"
	"eltimn/todo-plus/pkg/errs"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// pwd: 3ClyP69h9ax1zNE

func todoCollection() *mongo.Collection {
	return mainDB().Collection("todos")
}

type Todo struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id,omitempty"`
	PlainText   string             `bson:"plain_text,omitempty"`
	RichText    string             `bson:"rich_text,omitempty"`
	IsCompleted bool               `bson:"is_completed"`
}

func CreateNewTodo(cntxt context.Context, userId primitive.ObjectID, plainText, richText string) error {
	newTodo := Todo{
		UserID:      userId,
		PlainText:   plainText,
		RichText:    richText,
		IsCompleted: false,
	}

	slog.Info("newTodo", slog.Any("newTodo", newTodo))

	ctx, cancel := context.WithTimeout(cntxt, 5*time.Second)
	defer cancel()
	res, err := todoCollection().InsertOne(ctx, newTodo)
	if err != nil {
		slog.Error("Error creating new Todo", errs.ErrAttr(err))
		return err
	}
	id := res.InsertedID
	slog.Info("InsertedID", slog.Any("id", id))

	return nil
}

func DeleteTodoById(cntxt context.Context, todoId primitive.ObjectID) error {
	filter := bson.D{{Key: "_id", Value: todoId}}
	_, err := todoCollection().DeleteOne(cntxt, filter)
	if err != nil {
		slog.Error("Error deleting a Todo", errs.ErrAttr(err))
		return err
	}

	return nil
}

func FetchTodo(cntxt context.Context, todoId primitive.ObjectID) (Todo, error) {
	filter := bson.D{{Key: "_id", Value: todoId}}
	var result Todo
	err := todoCollection().FindOne(cntxt, filter).Decode(&result)
	if err != nil {
		slog.Error("Error fetching a Todo", errs.ErrAttr(err))
		return Todo{}, err
	}

	return result, nil
}

func FetchTodos(cntxt context.Context, userId primitive.ObjectID, filter string) ([]Todo, int, error) {
	// todos := []Todo{
	// 	{ID: primitive.NewObjectID(), UserID: userId, PlainText: "Buy groceries", RichText: "Buy groceries", IsCompleted: true},
	// 	{ID: primitive.NewObjectID(), UserID: userId, PlainText: "Call mom", RichText: "Call mom", IsCompleted: false},
	// 	{ID: primitive.NewObjectID(), UserID: userId, PlainText: "Write blog post", RichText: "Write blog post", IsCompleted: false},
	// }
	var bsonFilter bson.D
	switch filter {
	case "active":
		bsonFilter = bson.D{
			{
				Key: "$and",
				Value: bson.A{
					bson.D{{Key: "user_id", Value: userId}},
					bson.D{{Key: "is_completed", Value: false}},
				},
			},
		}
	case "completed":
		bsonFilter = bson.D{
			{
				Key: "$and",
				Value: bson.A{
					bson.D{{Key: "user_id", Value: userId}},
					bson.D{{Key: "is_completed", Value: true}},
				},
			},
		}
	default:
		bsonFilter = bson.D{{Key: "user_id", Value: userId}}
	}

	// Retrieves documents that match the query filer
	cursor, err := todoCollection().Find(cntxt, bsonFilter)
	if err != nil {
		slog.Error("Error fetching todos", errs.ErrAttr(err))
		return nil, 0, err
	}

	var results []Todo
	if err = cursor.All(cntxt, &results); err != nil {
		slog.Error("Error reading the todos cursor", errs.ErrAttr(err))
		return nil, 0, err
	}

	var activeCount int
	for i := range results {
		if !results[i].IsCompleted {
			activeCount++
		}
	}

	return results, activeCount, nil
}

func ToggleTodoCompleted(cntxt context.Context, todo Todo) error {
	filter := bson.D{{Key: "_id", Value: todo.ID}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "is_completed", Value: !todo.IsCompleted}}}}
	_, err := todoCollection().UpdateOne(cntxt, filter, update)
	if err != nil {
		slog.Error("Error updating todo", errs.ErrAttr(err))
		return err
	}

	return nil
}

func ToggleAllCompleted(cntxt context.Context, userId primitive.ObjectID, isCompleted bool) error {
	filter := bson.D{{Key: "user_id", Value: userId}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "is_completed", Value: isCompleted}}}}
	_, err := todoCollection().UpdateMany(cntxt, filter, update)
	if err != nil {
		slog.Error("Error updating all todos", errs.ErrAttr(err))
		return err
	}

	return nil
}

func DeleteAllCompleted(cntxt context.Context, userId primitive.ObjectID) error {
	filter := bson.D{{Key: "user_id", Value: userId}, {Key: "is_completed", Value: true}}
	_, err := todoCollection().DeleteMany(cntxt, filter)
	if err != nil {
		slog.Error("Error deleting all completed todos", errs.ErrAttr(err))
		return err
	}

	return nil
}
