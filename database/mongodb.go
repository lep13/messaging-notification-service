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
func InitializeMongoDB() {
	// Fetch secrets including MongoDB credentials
	secretName := "mongodbcreds"
	secrets, err := secretsmanager.GetSecretData(secretName)
	if err != nil {
		log.Fatalf("Failed to get secret data: %v", err)
	}

	// Use the fetched MongoDB credentials
	creds := secrets.MongoCredentials
	uri := creds.MongoDBURI
	clientOptions := options.Client().ApplyURI(uri)

	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Adding a timeout to the ping context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB successfully")
}

// GetCollection returns a collection from the MongoDB database
func GetCollection(collectionName string) *mongo.Collection {
	return client.Database("PROJECT0").Collection(collectionName)
}
