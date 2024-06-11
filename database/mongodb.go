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
	creds, err := secretsmanager.GetMongoCredentials("mongodbcreds")
	if err != nil {
		log.Fatalf("Failed to get MongoDB credentials: %v", err)
	}


	uri := "mongodb+srv://" + creds.Username + ":" + creds.Password + "@messages-cluster.utdovdn.mongodb.net/PROJECT0?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)

	client, err = mongo.Connect(context.TODO(), clientOptions)
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
