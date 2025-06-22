package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
}

func NewMongoDB(uri string) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(100) // Set the maximum pool size
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}
	return &MongoDB{client: client}, nil
}

func (m *MongoDB) GetCollection(databaseName string, collectionName string) *mongo.Collection {
	return m.client.Database(databaseName).Collection(collectionName)
}

func (m *MongoDB) Close() error {
	if err := m.client.Disconnect(context.Background()); err != nil {
		return err
	}
	return nil
}
