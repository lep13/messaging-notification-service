package main

import (
	"context"
	"log"

	"github.com/IBM/sarama"
	"github.com/lep13/messaging-notification-service/auth"
	"github.com/lep13/messaging-notification-service/database"
	"github.com/lep13/messaging-notification-service/kafka"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
	consumer, err := sarama.NewConsumer(addrs, config)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

func main() {
	// Create a new SecretManager instance
	secretManager := secretsmanager.NewSecretManager(nil)

	// Fetch secrets including Cognito token and MongoDB credentials
	secretName := "notifsecrets"
	secrets, err := secretManager.GetSecretData(secretName)
	if err != nil {
		log.Printf("Failed to get secret data: %v", err)
		return
	}

	cognitoToken := secrets.CognitoToken
	if cognitoToken == "" {
		log.Println("COGNITO_TOKEN not found in secrets manager")
		return
	}

	// Validate the Cognito token
	claims, err := auth.ValidateCognitoToken(context.Background(), cognitoToken, secretManager, auth.DefaultTokenValidator)
	if err != nil {
		log.Printf("Cognito token validation failed: %s", err)
		return
	}
	log.Printf("Cognito token validated: %v", claims)

	// Validate the "to" user using the profile token and mock API
	profile, err := auth.ValidateUserProfile(context.Background(), secretManager)
	if err != nil {
		log.Printf("Failed to validate user profile: %v", err)
		return
	}

	log.Printf("Validated user profile: %v", profile)

	// Define the dependencies for Kafka consumer
	deps := kafka.ConsumerDependencies{
		NewSecretManager: func(client secretsmanager.SecretsManagerAPI) secretsmanager.SecretManager {
			return secretManager
		},
		InitializeMongoDB: database.InitializeMongoDB,
		GetCollection: func(collectionName string) database.CollectionInterface {
			return database.GetCollection(collectionName)
		},
		NewKafkaConsumer: NewKafkaConsumer,
	}

	// Now continue with consuming messages from Kafka
	kafka.ConsumeMessages(deps)
}
