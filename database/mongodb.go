package database

import (
    "context"
    "log"
    "github.com/lep13/messaging-notification-service/secrets-manager"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Message represents the structure of the message document in MongoDB
type Message struct {
    FromUser           string `bson:"from_user"`
    ToUser             string `bson:"to_user"`
    MessageContent     string `bson:"message_content"`
    MessageConfirmation bool   `bson:"message_confirmation"`
}

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

// InsertMessage inserts a message document into the MongoDB collection
func InsertMessage(fromUser, toUser, messageContent string, messageConfirmation bool) error {
    collection := client.Database("PROJECT0").Collection("messages")

    msg := Message{
        FromUser:           fromUser,
        ToUser:             toUser,
        MessageContent:     messageContent,
        MessageConfirmation: messageConfirmation,
    }

    _, err := collection.InsertOne(context.TODO(), msg)
    if err != nil {
        return err
    }

    log.Printf("Inserted message from %s to %s into MongoDB", fromUser, toUser)
    return nil
}
