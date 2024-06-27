package runner

import (
	"context"
	"log"

	"github.com/IBM/sarama"
	"github.com/lep13/messaging-notification-service/auth"
	"github.com/lep13/messaging-notification-service/database"
	"github.com/lep13/messaging-notification-service/kafka"
	"github.com/lep13/messaging-notification-service/models"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

type RunnerDependencies struct {
	SecretManager     secretsmanager.SecretManager
	AuthValidator     auth.TokenValidatorFunc
	ProfileValidator  func(ctx context.Context, sm secretsmanager.SecretManager) (*models.Profile, error)
	InitializeMongoDB func() error
	GetCollection     func(collectionName string) database.CollectionInterface
	NewKafkaConsumer  func(addrs []string, config *sarama.Config) (sarama.Consumer, error)
}

func Run(deps RunnerDependencies) {
	// Fetch secrets including Cognito token and MongoDB credentials
	secretName := "notifsecrets"
	secrets, err := deps.SecretManager.GetSecretData(secretName)
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
	claims, err := auth.ValidateCognitoToken(context.Background(), cognitoToken, deps.SecretManager, deps.AuthValidator)
	if err != nil {
		log.Printf("Cognito token validation failed: %s", err)
		return
	}
	log.Printf("Cognito token validated: %v", claims)

	// Validate the "to" user using the profile token and mock API
	profile, err := deps.ProfileValidator(context.Background(), deps.SecretManager)
	if err != nil {
		log.Printf("Failed to validate user profile: %v", err)
		return
	}

	log.Printf("Validated user profile: %v", profile)

	// Initialize MongoDB
	err = deps.InitializeMongoDB()
	if err != nil {
		log.Printf("Failed to initialize MongoDB: %v", err)
		panic(err)
	}

	// Define the dependencies for Kafka consumer
	consumerDeps := kafka.ConsumerDependencies{
		NewSecretManager: func(client secretsmanager.SecretsManagerAPI) secretsmanager.SecretManager {
			return deps.SecretManager
		},
		InitializeMongoDB: deps.InitializeMongoDB,
		GetCollection: func(collectionName string) database.CollectionInterface {
			return deps.GetCollection(collectionName)
		},
		NewKafkaConsumer: deps.NewKafkaConsumer,
	}

	// Now continue with consuming messages from Kafka
	kafka.ConsumeMessages(consumerDeps)
}

func DefaultRun() {
	secretManager := secretsmanager.NewSecretManager(nil)
	deps := RunnerDependencies{
		SecretManager:     secretManager,
		AuthValidator:     auth.DefaultTokenValidator,
		ProfileValidator:  auth.ValidateUserProfile,
		InitializeMongoDB: database.InitializeMongoDB,
		GetCollection: func(collectionName string) database.CollectionInterface {
			return database.GetCollection(collectionName)
		},
		NewKafkaConsumer: NewKafkaConsumer,
	}
	Run(deps)
}

func NewKafkaConsumer(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
	consumer, err := sarama.NewConsumer(addrs, config)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}
