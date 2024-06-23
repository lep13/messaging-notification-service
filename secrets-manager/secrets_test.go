package secretsmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
)

// MockSecretManagerClient simulates the behavior of the AWS Secrets Manager client
type MockSecretManagerClient struct {
	GetSecretValueFunc func(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func (m *MockSecretManagerClient) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m.GetSecretValueFunc(ctx, input, opts...)
}

// Mock function to load AWS configuration
func mockLoadAWSConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return aws.Config{Region: "us-east-1"}, nil
}

func TestGetSecretData(t *testing.T) {
	// Mock the AWS configuration loading function
	originalLoadAWSConfig := loadAWSConfig
	defer func() { loadAWSConfig = originalLoadAWSConfig }()

	// Mock successful AWS config loading for tests where we are not testing config loading failure
	loadAWSConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{Region: "us-east-1"}, nil
	}

	t.Run("Successful Secret Retrieval", func(t *testing.T) {
		mockClient := &MockSecretManagerClient{
			GetSecretValueFunc: func(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				secretString := `{"mongodbURI": "mongodb://test:27017", "COGNITO_TOKEN": "some_token"}`
				return &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(secretString),
				}, nil
			},
		}

		sm := NewSecretManager(mockClient)

		secretData, err := sm.GetSecretData("test-secret")
		assert.NoError(t, err)
		assert.Equal(t, "mongodb://test:27017", secretData.MongoDBURI)
		assert.Equal(t, "some_token", secretData.CognitoToken)
	})

	t.Run("AWS Configuration Loading Failure", func(t *testing.T) {
		// Test specifically for AWS config loading failure
		loadAWSConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
			return aws.Config{}, errors.New("failed to load AWS config")
		}

		sm := NewSecretManager(nil)

		secretData, err := sm.GetSecretData("test-secret")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load AWS config")
		assert.Empty(t, secretData)
	})

	t.Run("Secret Not Found", func(t *testing.T) {
		// Restore AWS config loading to succeed for this test
		loadAWSConfig = mockLoadAWSConfig

		mockClient := &MockSecretManagerClient{
			GetSecretValueFunc: func(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				return nil, errors.New("the requested secret test-secret was not found")
			},
		}

		sm := NewSecretManager(mockClient)

		secretData, err := sm.GetSecretData("test-secret")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the requested secret test-secret was not found")
		assert.Empty(t, secretData)
	})

	t.Run("Failed to Retrieve Secret Value", func(t *testing.T) {
		// Ensure AWS config loading succeeds for this test
		loadAWSConfig = mockLoadAWSConfig

		mockClient := &MockSecretManagerClient{
			GetSecretValueFunc: func(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				return nil, errors.New("failed to retrieve secret value")
			},
		}

		sm := NewSecretManager(mockClient)

		secretData, err := sm.GetSecretData("test-secret")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve secret value")
		assert.Empty(t, secretData)
	})

	t.Run("Nil Secret String", func(t *testing.T) {
		// Ensure AWS config loading succeeds for this test
		loadAWSConfig = mockLoadAWSConfig

		mockClient := &MockSecretManagerClient{
			GetSecretValueFunc: func(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				return &secretsmanager.GetSecretValueOutput{
					SecretString: nil,
				}, nil
			},
		}

		sm := NewSecretManager(mockClient)

		secretData, err := sm.GetSecretData("test-secret")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret string is nil")
		assert.Empty(t, secretData)
	})

	t.Run("Failed to Unmarshal Secret String", func(t *testing.T) {
		// Ensure AWS config loading succeeds for this test
		loadAWSConfig = mockLoadAWSConfig

		mockClient := &MockSecretManagerClient{
			GetSecretValueFunc: func(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				secretString := `{"invalid_json"}`
				return &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(secretString),
				}, nil
			},
		}

		sm := NewSecretManager(mockClient)

		secretData, err := sm.GetSecretData("test-secret")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal secret string")
		assert.Empty(t, secretData)
	})
}
