package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/ShreerajShettyK/cognitoJwtAuthenticator"
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

func TestValidateCognitoToken_ErrorRetrievingSecrets(t *testing.T) {
	ctx := context.Background()
	tokenString := "valid-token"

	mockSM := &MockSecretsManagerWithError{
		err: errors.New("error retrieving secrets"),
	}

	claims, err := ValidateCognitoToken(ctx, tokenString, mockSM, MockValidateToken)

	assert.NotNil(t, err)
	assert.Nil(t, claims)
}

func TestValidateCognitoToken_ErrorValidatingToken(t *testing.T) {
	ctx := context.Background()
	tokenString := "invalid-token"

	mockSM := &MockSecretsManager{
		Region:     "us-east-1",
		UserPoolID: "us-east-1_abc123",
	}

	claims, err := ValidateCognitoToken(ctx, tokenString, mockSM, MockValidateToken)

	assert.NotNil(t, err)
	assert.Nil(t, claims)
}
