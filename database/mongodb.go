package database

import (
	"context"
	"log"
	"time"

	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// InitializeMongoDB initializes the MongoDB client connection
func InitializeMongoDB() error {
	// Create a secret manager instance
	secretManager := secretsmanager.NewSecretManager(nil)

	// Fetch secrets including MongoDB URI
	secretName := "notifsecrets"
	secrets, err := secretManager.GetSecretData(secretName)
	if err != nil {
		log.Printf("Failed to get secret data: %v", err)
		return err
	}

	// Use the MongoDB URI from the secrets
	uri := secrets.MongoDBURI

	clientOptions := options.Client().ApplyURI(uri)

	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return err
	}

	// Adding a timeout to the ping context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		return err
	}

	log.Println("Connected to MongoDB successfully")
	return nil
}

// GetCollection returns a collection from the MongoDB database
func GetCollection(collectionName string) *mongo.Collection {
	return client.Database("PROJECT0").Collection(collectionName)
}