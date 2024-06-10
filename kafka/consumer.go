package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/lep13/messaging-notification-service/auth"
	"github.com/lep13/messaging-notification-service/database"
	"github.com/segmentio/kafka-go"
)

// Message represents the structure of the Kafka message
type Message struct {
	FromUser string `json:"from_user"`
	ToUser   string `json:"to_user"`
	Content  string `json:"content"`
}

func ConsumeMessages(ctx context.Context, reader *kafka.Reader) {
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("error while reading message: %v", err)
			continue
		}

		var message Message
		err = json.Unmarshal(msg.Value, &message)
		if err != nil {
			log.Printf("error unmarshalling message: %v", err)
			continue
		}

		log.Printf("received message: %+v", message)

		// Validate "To" user
		isValid, err := auth.ValidateUser(message.ToUser)
		if err != nil || !isValid {
			log.Printf("error validating user: %v", err)
			continue
		}

		// Store the message in MongoDB
		err = database.InsertMessage(message.FromUser, message.ToUser, message.Content, false)
		if err != nil {
			log.Printf("error inserting message into MongoDB: %v", err)
		}
	}
}
