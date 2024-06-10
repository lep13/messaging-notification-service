package database

import (
    "context"
    "log"
    "github.com/lep13/messaging-notification-service/secrets-manager" 
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// InitializeMongoDB initializes the MongoDB client connection
func InitializeMongoDB() {
	creds, err := secretsmanager.GetMongoCredentials("mongo/credentials")
	if err != nil {
		log.Fatalf("Failed to get MongoDB credentials: %v", err)
	}

	uri := "mongodb+srv://" + creds.Username + ":" + creds.Password + "@messages-cluster.mongodb.net/PROJECT0?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)

	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB successfully")
}
