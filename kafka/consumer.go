package kafka

import (
	"context"
	"log"
	"regexp"

	"github.com/IBM/sarama"
	"github.com/lep13/messaging-notification-service/database"
	"go.mongodb.org/mongo-driver/bson"
)

// ConsumeMessages initializes Kafka consumer and handles messages
func ConsumeMessages() {
	log.Println("Starting Kafka consumer...")

	// Kafka consumer configuration
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Version = sarama.V2_1_0_0 // Ensure you are using a compatible version

	// Creating a new Kafka consumer
	consumer, err := sarama.NewConsumer([]string{"34.224.79.8:9092"}, config)
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
	partitionConsumer, err := consumer.ConsumePartition("chat-topic-46", 0, sarama.OffsetOldest)
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

	// Assuming the message is in the format "From:<from>, To:<to>, Message:<msg>"
	parsedMessage := parseMessage(string(value))
	if parsedMessage == nil {
		log.Printf("Failed to parse message: %s", string(value))
		return
	}

	// Insert the document into MongoDB
	collection := database.GetCollection("messages")
	_, err := collection.InsertOne(context.TODO(), parsedMessage)
	if err != nil {
		log.Printf("Failed to insert document into MongoDB: %v", err)
		return
	}

	log.Printf("Message inserted into MongoDB: %s", string(value))
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
