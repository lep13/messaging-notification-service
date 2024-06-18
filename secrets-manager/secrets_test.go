package secretsmanager

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/lep13/messaging-notification-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSecretData(t *testing.T) {
	// Define the mock responses for different scenarios
	mockSecretData := models.SecretData{
		MongoDBURI:           "mongodb+srv://lepakshi1304:juSXfZFfMT3UWqKh@messages-cluster.utdovdn.mongodb.net/PROJECT0?retryWrites=true&w=majority",
		CognitoToken:         "mock_token",
		Region:               "us-east-1",
		UserPoolID:           "us-east-1_icXeg2eiv",
		ProfileURL:           "https://fc76bbfd-f022-43e1-bb9d-355a53bae2cc.mock.pstmn.io",
		KafkaBroker:          "54.226.200.228:9092",
		KafkaTopic:           "chat-topic-77",
		NotificationEndpoint: "http://localhost:8081/notify",
	}

	secretString, _ := json.Marshal(mockSecretData)

	// Mock function for successful response
	mockGetSecretValue := func(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		return &secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(string(secretString)),
		}, nil
	}

	// Test cases
	t.Run("Successful retrieval of all secrets", func(t *testing.T) {
		secretName := "notifsecrets"
		secretData, err := GetSecretData(secretName, mockGetSecretValue)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if secretData.MongoDBURI != mockSecretData.MongoDBURI {
			t.Fatalf("Expected MongoDB URI %s, got %s", mockSecretData.MongoDBURI, secretData.MongoDBURI)
		}
		if secretData.CognitoToken != mockSecretData.CognitoToken {
			t.Fatalf("Expected Cognito Token %s, got %s", mockSecretData.CognitoToken, secretData.CognitoToken)
		}
		if secretData.Region != mockSecretData.Region {
			t.Fatalf("Expected Region %s, got %s", mockSecretData.Region, secretData.Region)
		}
		if secretData.UserPoolID != mockSecretData.UserPoolID {
			t.Fatalf("Expected User Pool ID %s, got %s", mockSecretData.UserPoolID, secretData.UserPoolID)
		}
		if secretData.ProfileURL != mockSecretData.ProfileURL {
			t.Fatalf("Expected Profile URL %s, got %s", mockSecretData.ProfileURL, secretData.ProfileURL)
		}
		if secretData.KafkaBroker != mockSecretData.KafkaBroker {
			t.Fatalf("Expected Kafka Broker %s, got %s", mockSecretData.KafkaBroker, secretData.KafkaBroker)
		}
		if secretData.KafkaTopic != mockSecretData.KafkaTopic {
			t.Fatalf("Expected Kafka Topic %s, got %s", mockSecretData.KafkaTopic, secretData.KafkaTopic)
		}
		if secretData.NotificationEndpoint != mockSecretData.NotificationEndpoint {
			t.Fatalf("Expected Notification Endpoint %s, got %s", mockSecretData.NotificationEndpoint, secretData.NotificationEndpoint)
		}
	})

	t.Run("Secret not found in AWS Secrets Manager", func(t *testing.T) {
		mockGetSecretValue = func(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return nil, errors.New("ResourceNotFoundException: the requested secret testsecret was not found")
		}

		_, err := GetSecretData("testsecret", mockGetSecretValue)
		if err == nil || !strings.Contains(err.Error(), "the requested secret testsecret was not found") {
			t.Fatalf("Expected 'the requested secret testsecret was not found' error, got %v", err)
		}
	})

	t.Run("Failed to retrieve secret value", func(t *testing.T) {
		mockGetSecretValue = func(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return nil, errors.New("failed to retrieve secret value")
		}

		_, err := GetSecretData("testsecret", mockGetSecretValue)
		if err == nil || !strings.Contains(err.Error(), "failed to retrieve secret value") {
			t.Fatalf("Expected 'failed to retrieve secret value' error, got %v", err)
		}
	})

	t.Run("Secret string is nil", func(t *testing.T) {
		mockGetSecretValue = func(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{
				SecretString: nil,
			}, nil
		}

		_, err := GetSecretData("testsecret", mockGetSecretValue)
		if err == nil || !strings.Contains(err.Error(), "secret string is nil") {
			t.Fatalf("Expected 'secret string is nil' error, got %v", err)
		}
	})

	t.Run("Failed to unmarshal secret string", func(t *testing.T) {
		// Create a JSON string with an invalid field type for `models.SecretData`
		badSecretString := `{"MongoDBURI": {"invalid": "type"}}`
		mockGetSecretValue = func(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String(badSecretString),
			}, nil
		}

		_, err := GetSecretData("testsecret", mockGetSecretValue)
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal secret string") {
			t.Fatalf("Expected 'failed to unmarshal secret string' error, got %v", err)
		}
	})
}
func TestNewSecretsManagerClient(t *testing.T) {
	t.Run("Successful creation of Secrets Manager client", func(t *testing.T) {
		// Use the actual config.LoadDefaultConfig function for successful test
		client, err := NewSecretsManagerClient("us-east-1", config.LoadDefaultConfig)
		require.NoError(t, err, "Expected no error in creating the Secrets Manager client")
		require.NotNil(t, client, "Expected non-nil Secrets Manager client")
	})

	t.Run("Error loading AWS configuration", func(t *testing.T) {
		// Mock function to simulate error during config load
		mockLoadConfigFunc := func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
			return aws.Config{}, errors.New("mock error loading configuration")
		}

		client, err := NewSecretsManagerClient("us-east-1", mockLoadConfigFunc)
		assert.Error(t, err, "Expected error in loading AWS configuration")
		assert.Nil(t, client, "Expected nil Secrets Manager client on error")
	})
}
