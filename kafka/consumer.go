package kafka

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/IBM/sarama"
	"github.com/lep13/messaging-notification-service/database"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
	"github.com/lep13/messaging-notification-service/services"
	"go.mongodb.org/mongo-driver/bson"
)

// ConsumeMessages initializes Kafka consumer and handles messages
func ConsumeMessages() {
	log.Println("Starting Kafka consumer...")

	// Fetch secrets including Kafka broker IP and topic
	secretName := "notifsecrets"
	secrets, err := secretsmanager.GetSecretData(secretName)
	if err != nil {
		log.Fatalf("Error retrieving secrets: %v", err)
		return
	}

	kafkaBroker := secrets.KafkaBroker
	kafkaTopic := secrets.KafkaTopic

	// Kafka consumer configuration
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Version = sarama.V2_1_0_0

	// Creating a new Kafka consumer
	consumer, err := sarama.NewConsumer([]string{kafkaBroker}, config)
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
		return
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalf("Failed to close consumer: %v", err)
		}
	}()

	// Consume from the specified partition
	partitionConsumer, err := consumer.ConsumePartition(kafkaTopic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatalf("Failed to start partition consumer: %v", err)
		return
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalf("Failed to close partition consumer: %v", err)
		}
	}()

	log.Println("Kafka consumer started successfully")
	log.Println("Partition consumer started. Waiting for messages...")

	// Initialize MongoDB connection
	database.InitializeMongoDB()

	// Process messages
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Received message from Kafka: %s", string(msg.Value))
			processMessage(msg.Value)

		case err := <-partitionConsumer.Errors():
			log.Printf("Error consuming message: %v", err)
		}
	}
}

// processMessage handles the received Kafka message
func processMessage(value []byte) {
	log.Printf("Processing message: %s", string(value))

	// message is in the format "From:<from>, To:<to>, Message:<msg>"
	parsedMessage := parseMessage(string(value))
	if parsedMessage == nil {
		log.Printf("Failed to parse message: %s", string(value))
		return
	}

	// Insert the document into MongoDB
	collection := database.GetCollection("messages")
	insertResult, err := collection.InsertOne(context.TODO(), parsedMessage)
	if err != nil {
		log.Printf("Failed to insert document into MongoDB: %v", err)
		return
	}

	log.Printf("Message inserted into MongoDB: %s", string(value))

	// Prepare the notification
	notification := services.Notification{
		From:    parsedMessage["from"].(string),
		To:      parsedMessage["to"].(string),
		Message: parsedMessage["message"].(string),
	}

	// Notify the UI about the new message
	token := "example_token" // Replace with the actual token logic as per your setup
	err = services.NotifyUI(notification, token)
	if err != nil {
		log.Printf("Failed to send notification: %v", err)

		// Update the document with notified status false and reason for failure
		update := bson.M{
			"$set": bson.M{
				"notified": false,
				"reason":   "Failed to send notification",
			},
		}
		_, updateErr := collection.UpdateByID(context.Background(), insertResult.InsertedID, update)
		if updateErr != nil {
			log.Printf("Failed to update document with failure reason: %v", updateErr)
		}
		return
	}

	// After notification, update the document with a notified status
	update := bson.M{
		"$set": bson.M{
			"notified":   true,
			"notifiedAt": time.Now(),
		},
	}
	_, err = collection.UpdateByID(context.Background(), insertResult.InsertedID, update)
	if err != nil {
		log.Printf("Failed to update document with notified status: %v", err)
	}
}

// parseMessage parses a Kafka message and returns a BSON document
func parseMessage(message string) bson.M {
	// Use regular expressions to capture From, To, and Message parts
	re := regexp.MustCompile(`From:(.*?), To:(.*?), Message:(.*)`)
	matches := re.FindStringSubmatch(message)

	if len(matches) != 4 {
		return nil
	}

	return bson.M{
		"from":    matches[1],
		"to":      matches[2],
		"message": matches[3],
	}
}
