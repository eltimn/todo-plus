package models

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// https://www.alexedwards.net/blog/organising-database-access
var client *mongo.Client
var ctx = context.TODO()

func mainDB() *mongo.Database {
	return client.Database("todo_plus")
}

func InitMongoDB(uri string) error {
	var err error

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	return client.Ping(ctx, nil)
}

func ShutdownMongoDB() error {
	err := client.Disconnect(ctx)
	if err != nil {
		return err
	}

	slog.Info("Disconnected from MongoDB")
	return nil
}
