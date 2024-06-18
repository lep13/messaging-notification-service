package secretsmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/lep13/messaging-notification-service/models"
)

// GetSecretData fetches secrets from AWS Secrets Manager
func GetSecretData(secretName string, getSecretValueFunc func(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)) (models.SecretData, error) {
	var secretData models.SecretData

	// Retrieve the secret value using the provided function
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := getSecretValueFunc(context.TODO(), input)
	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
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

// NewSecretsManagerClient creates a new AWS Secrets Manager client
func NewSecretsManagerClient(region string, loadConfigFunc func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error)) (*secretsmanager.Client, error) {
	// Load the AWS configuration using the provided function and region
	cfg, err := loadConfigFunc(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	// Create a Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)
	return client, nil
}
