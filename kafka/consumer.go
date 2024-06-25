package kafka

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/IBM/sarama"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/lep13/messaging-notification-service/database"
	"github.com/lep13/messaging-notification-service/models"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
	"github.com/lep13/messaging-notification-service/services"
)

// Dependencies for the ConsumeMessages function
type ConsumerDependencies struct {
	NewSecretManager  func(client secretsmanager.SecretsManagerAPI) secretsmanager.SecretManager
	InitializeMongoDB func() error
	GetCollection     func(collectionName string) database.CollectionInterface
	NewKafkaConsumer  func(addrs []string, config *sarama.Config) (sarama.Consumer, error)
}

// ConsumeMessages initializes Kafka consumer and handles messages.
func ConsumeMessages(deps ConsumerDependencies) {
	log.Println("Starting Kafka consumer...")

	// Create a secret manager instance
	secretManager := deps.NewSecretManager(nil)
	if secretManager == nil {
		log.Println("Failed to create secret manager instance")
		return
	}
	log.Println("Secret manager created successfully")

	// Fetch secrets including Kafka broker IP and topic
	secretName := "notifsecrets"
	secrets, err := secretManager.GetSecretData(secretName)
	if err != nil {
		log.Printf("Error retrieving secrets: %v", err)
		return
	}

	kafkaBroker := secrets.KafkaBroker
	kafkaTopic := secrets.KafkaTopic

	log.Printf("Kafka broker: %s, Kafka topic: %s", kafkaBroker, kafkaTopic)

	// Kafka consumer configuration
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Version = sarama.V2_1_0_0

	// Creating a new Kafka consumer
	consumer, err := deps.NewKafkaConsumer([]string{kafkaBroker}, config)
	if err != nil {
		log.Printf("Failed to start consumer: %v", err)
		return
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Failed to close consumer: %v", err)
		}
	}()
	log.Println("Kafka consumer created successfully")

	// Consume from the specified partition
	partitionConsumer, err := consumer.ConsumePartition(kafkaTopic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Printf("Failed to start partition consumer: %v", err)
		return
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Failed to close partition consumer: %v", err)
		}
	}()
	log.Println("Partition consumer started successfully. Waiting for messages...")

	// Initialize MongoDB connection
	if err := deps.InitializeMongoDB(); err != nil {
		log.Printf("Failed to initialize MongoDB: %v", err)
		return
	}
	log.Println("MongoDB initialized successfully")

	// Wrapper function for GetCollection to match expected signature
	getCollectionWrapper := func(collectionName string) database.CollectionInterface {
		return deps.GetCollection(collectionName)
	}

	// Wrapper for NotifyUI to adapt to the expected signature
	notifyUIWrapper := func(notification models.Notification, token string, secretManager secretsmanager.SecretManager) error {
		return services.NotifyUI(notification, token, secretManager)
	}

	// Process messages
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Received message from Kafka: %s", string(msg.Value))
			processMessage(msg.Value, secretManager, getCollectionWrapper, notifyUIWrapper)

		case err := <-partitionConsumer.Errors():
			log.Printf("Error consuming message: %v", err)
		}
	}
}

// processMessage handles the received Kafka message.
func processMessage(value []byte, secretManager secretsmanager.SecretManager, getCollection func(string) database.CollectionInterface, notifyUI func(models.Notification, string, secretsmanager.SecretManager) error) {
	log.Printf("Processing message: %s", string(value))

	parsedMessage := parseMessage(string(value))
	if parsedMessage == nil {
		log.Printf("Failed to parse message: %s", string(value))
		return
	}

	collection := getCollection("messages")
	insertResult, err := collection.InsertOne(context.TODO(), parsedMessage)
	if err != nil {
		log.Printf("Failed to insert document into MongoDB: %v", err)
		return
	}

	log.Printf("Message inserted into MongoDB: %s", string(value))

	notification := models.Notification{
		From:    parsedMessage["from"].(string),
		To:      parsedMessage["to"].(string),
		Message: parsedMessage["message"].(string),
	}

	token := "example_token" // Replace with actual token logic as per your setup
	err = notifyUI(notification, token, secretManager)
	if err != nil {
		log.Printf("Failed to send notification: %v", err)

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

// parseMessage parses a Kafka message and returns a BSON document.
func parseMessage(message string) bson.M {
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
