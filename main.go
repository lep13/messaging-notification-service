package main

import (
	"context"
	"log"

	"github.com/lep13/messaging-notification-service/auth"
	"github.com/lep13/messaging-notification-service/kafka"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

func main() {
	// Fetch secrets 
	secretName := "notifsecrets"
	secrets, err := secretsmanager.GetSecretData(secretName)
	if err != nil {
		log.Fatalf("Failed to get secret data: %v", err)
	}

	cognitoToken := secrets.CognitoToken
	if cognitoToken == "" {
		log.Fatal("COGNITO_TOKEN not found in secrets manager")
	}

	// Validate the Cognito token
	claims, err := auth.ValidateCognitoToken(context.Background(), cognitoToken)
	if err != nil {
		log.Fatalf("Cognito token validation failed: %s", err)
	}
	log.Printf("Cognito token validated: %v", claims)

	// Validate the "to" user using the profile token and mock API
	profile, err := auth.ValidateUserProfile(context.Background())
	if err != nil {
		log.Fatalf("Failed to validate user profile: %v", err)
	}

	log.Printf("Validated user profile: %v", profile)

	// Now continue with consuming messages from Kafka
	kafka.ConsumeMessages()
}
