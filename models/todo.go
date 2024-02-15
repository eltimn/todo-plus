package models

import (
	"context"
	"eltimn/todo-plus/utils"
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

func FetchTodos(userId primitive.ObjectID) ([]Todo, int, error) {
	// todos := []Todo{
	// 	{ID: primitive.NewObjectID(), UserID: userId, PlainText: "Buy groceries", RichText: "Buy groceries", IsCompleted: true},
	// 	{ID: primitive.NewObjectID(), UserID: userId, PlainText: "Call mom", RichText: "Call mom", IsCompleted: false},
	// 	{ID: primitive.NewObjectID(), UserID: userId, PlainText: "Write blog post", RichText: "Write blog post", IsCompleted: false},
	// }

	filter := bson.D{{Key: "user_id", Value: userId}}
	// Retrieves documents that match the query filer
	cursor, err := todoCollection().Find(context.TODO(), filter)
	if err != nil {
		slog.Error("Error fetching todos", utils.ErrAttr(err))
	}

	var results []Todo
	if err = cursor.All(context.TODO(), &results); err != nil {
		slog.Error("Error reading the todos cursor", utils.ErrAttr(err))
	}

	var activeCount int
	for i := range results {
		if results[i].IsCompleted {
			activeCount++
		}
	}

	// fmt.Println("results", results)

	return results, activeCount, nil
}

func CreateNewTodo(userId primitive.ObjectID, plainText, richText string) error {
	newTodo := Todo{
		UserID:      userId,
		PlainText:   plainText,
		RichText:    richText,
		IsCompleted: false,
	}

	slog.Info("newTodo", slog.Any("newTodo", newTodo))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := todoCollection().InsertOne(ctx, newTodo)
	if err != nil {
		return err
	}
	id := res.InsertedID
	slog.Info("InsertedID", slog.Any("id", id))

	return nil
}
