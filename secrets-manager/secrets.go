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

// SecretsManagerAPI defines the methods that our AWS client will implement
type SecretsManagerAPI interface {
	GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// secretManager is our custom implementation that uses the AWS client
type secretManager struct {
	client SecretsManagerAPI
}

// SecretManager is the interface that our secret manager implements
type SecretManager interface {
	GetSecretData(secretName string) (models.SecretData, error)
}

// loadAWSConfig is a variable that points to the function that loads AWS config.
// This allows us to replace it with a mock in tests.
var loadAWSConfig = config.LoadDefaultConfig

// NewSecretManager creates a new instance of secretManager with the given client.
func NewSecretManager(client SecretsManagerAPI) *secretManager {
	return &secretManager{
		client: client,
	}
}

// GetSecretData fetches secrets from AWS Secrets Manager
func (sm *secretManager) GetSecretData(secretName string) (models.SecretData, error) {
	var secretData models.SecretData

	cfg, err := loadAWSConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return secretData, fmt.Errorf("unable to load SDK config, %v", err)
	}

	if sm.client == nil {
		sm.client = secretsmanager.NewFromConfig(cfg)
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := sm.client.GetSecretValue(context.TODO(), input)
	if err != nil {
		var re *types.ResourceNotFoundException
		if errors.As(err, &re) {
			return secretData, fmt.Errorf("the requested secret %s was not found", secretName)
		}
		return secretData, fmt.Errorf("failed to retrieve secret value, %v", err)
	}

	if result.SecretString == nil {
		return secretData, errors.New("secret string is nil")
	}

	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		return secretData, fmt.Errorf("failed to unmarshal secret string, %v", err)
	}

	return secretData, nil
}
