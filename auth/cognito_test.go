package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/ShreerajShettyK/cognitoJwtAuthenticator"
	// "github.com/lep13/messaging-notification-service/models"
	"github.com/stretchr/testify/assert"
)

func MockValidateToken(ctx context.Context, region, userPoolID, tokenString string) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
	if tokenString == "valid-token" {
		return &cognitoJwtAuthenticator.AWSCognitoClaims{}, nil
	}
	return nil, errors.New("invalid token")
}

func TestValidateCognitoToken(t *testing.T) {
	ctx := context.Background()
	tokenString := "valid-token"

	// Using the shared mock secrets manager with required fields for this test
	mockSM := &MockSecretsManager{
		Region:     "us-east-1",
		UserPoolID: "us-east-1_abc123",
	}

	// Injecting mock token validator into ValidateCognitoToken
	claims, err := ValidateCognitoToken(ctx, tokenString, mockSM, MockValidateToken)

	assert.Nil(t, err)
	assert.NotNil(t, claims)
}

// func TestValidateCognitoToken_ErrorFetchingSecrets(t *testing.T) {
// 	ctx := context.Background()
// 	tokenString := "valid-token"

// 	// Mock secrets manager to simulate error
// 	mockSM := &MockSecretsManager{
// 		Region:     "us-east-1",
// 		UserPoolID: "us-east-1_abc123",
// 	}
// 	// Override GetSecretData to return an error
// 	// originalGetSecretData := mockSM.GetSecretData
// 	// mockSM.GetSecretData = func(secretName string) (models.SecretData, error) {
// 	// 	return models.SecretData{}, errors.New("error fetching secrets")
// 	// }
// 	// defer func() { mockSM.GetSecretData = originalGetSecretData }() // Restore original function after test

// 	claims, err := ValidateCognitoToken(ctx, tokenString, mockSM, MockValidateToken)

// 	assert.NotNil(t, err)
// 	assert.Nil(t, claims)
// }

func TestValidateCognitoToken_TokenValidationError(t *testing.T) {
	ctx := context.Background()
	tokenString := "invalid-token"

	// Using the shared mock secrets manager
	mockSM := &MockSecretsManager{
		Region:     "us-east-1",
		UserPoolID: "us-east-1_abc123",
	}

	claims, err := ValidateCognitoToken(ctx, tokenString, mockSM, MockValidateToken)

	assert.NotNil(t, err)
	assert.Nil(t, claims)
}
