package main

import (
	"context"
	"log"

	"github.com/lep13/messaging-notification-service/auth"
	"github.com/lep13/messaging-notification-service/kafka"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

func main() {
	// Fetch secrets including Cognito token and MongoDB credentials
	secretName := "notifsecrets"
	secrets, err := secretsmanager.GetSecretData(secretName)
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
	claims, err := auth.ValidateCognitoToken(context.Background(), cognitoToken)
	if err != nil {
		log.Printf("Cognito token validation failed: %s", err)
		return
	}
	log.Printf("Cognito token validated: %v", claims)

	// Validate the "to" user using the profile token and mock API
	profile, err := auth.ValidateUserProfile(context.Background())
	if err != nil {
		log.Printf("Failed to validate user profile: %v", err)
		return
	}

	log.Printf("Validated user profile: %v", profile)

	// Now continue with consuming messages from Kafka
	kafka.ConsumeMessages()
}
