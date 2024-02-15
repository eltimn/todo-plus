package models

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// https://www.alexedwards.net/blog/organising-database-access
var client *mongo.Client
var mongoCtx context.Context

func mainDB() *mongo.Database {
	return client.Database("todo")
}

func InitMongoDB(uri string) error {
	var err error

	mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err = mongo.Connect(mongoCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	return client.Ping(mongoCtx, readpref.Primary())
}

func ShutdownMongoDB() error {
	err := client.Disconnect(mongoCtx)
	if err != nil {
		return err
	}

	slog.Info("Disconnected from MongoDB")
	return nil
}
