package auth

import (
	"context"
	"log"

	"github.com/ShreerajShettyK/cognitoJwtAuthenticator"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

// ValidateCognitoToken validates the provided JWT token with AWS Cognito using cognitoJwtAuthenticator
func ValidateCognitoToken(ctx context.Context, tokenString string) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
	// Fetch the region and userPoolId from secrets manager
	secretName := "notifsecrets"
	secrets, err := secretsmanager.GetSecretData(secretName)
	if err != nil {
		log.Printf("Error retrieving secrets: %v", err)
		return nil, err
	}

	region := secrets.Region
	userPoolId := secrets.UserPoolID

	claims, err := cognitoJwtAuthenticator.ValidateToken(ctx, region, userPoolId, tokenString)
	if err != nil {
		log.Printf("Token validation error: %s", err)
		return nil, err
	}
	log.Println("Token is valid")
	log.Printf("Claims: %+v", claims)
	return claims, nil
}
