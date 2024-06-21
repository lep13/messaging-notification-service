package secretsmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/lep13/messaging-notification-service/models"
)

type SecretManager interface {
    GetSecretData(secretName string) (models.SecretData, error)
}

// GetSecretData function adheres to this interface
var DefaultSecretManager SecretManager = &secretManager{}

type secretManager struct{}

// GetSecretData fetches secrets from AWS Secrets Manager
func (sm *secretManager) GetSecretData(secretName string) (models.SecretData, error) {
	var secretData models.SecretData

	// Load the AWS configuration from environment variables or the shared configuration file
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return secretData, fmt.Errorf("unable to load SDK config, %v", err)
	}

	// Create a Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Retrieve the secret value
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		var re *types.ResourceNotFoundException
		if errors.As(err, &re) {
			return secretData, fmt.Errorf("the requested secret %s was not found", secretName)
		}
		return secretData, fmt.Errorf("failed to retrieve secret value, %v", err)
	}

	// Check if the secret value is nil
	if result.SecretString == nil {
		return secretData, errors.New("secret string is nil")
	}

	// Unmarshal the secret string into the secretData struct
	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		return secretData, fmt.Errorf("failed to unmarshal secret string, %v", err)
	}

	return secretData, nil
}
