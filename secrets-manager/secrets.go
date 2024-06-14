package secretsmanager

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// MongoCredentials represents MongoDB credentials structure
type MongoCredentials struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	MongoDBURI string `json:"mongodbURI"` // Added field for MongoDB URI
}

// SecretData represents the structure for all secrets including MongoDB credentials and Cognito token
type SecretData struct {
	MongoCredentials     MongoCredentials `json:"mongodbcreds"`
	CognitoToken         string           `json:"COGNITO_TOKEN"`
	Region               string           `json:"region"`
	UserPoolID           string           `json:"userPoolId"`
	ProfileURL           string           `json:"profileURL"`
	KafkaBroker          string           `json:"kafkaBroker"`
	KafkaTopic           string           `json:"kafkaTopic"`
	NotificationEndpoint string           `json:"notificationEndpoint"`
}

// GetSecretData fetches from AWS Secrets Manager
func GetSecretData(secretName string) (SecretData, error) {
	var secretData SecretData

	sess, err := session.NewSession()
	if err != nil {
		return secretData, err
	}

	svc := secretsmanager.New(sess, aws.NewConfig().WithRegion("us-east-1"))

	result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return secretData, err
	}

	if result.SecretString == nil {
		return secretData, errors.New("secret string is nil")
	}

	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		return secretData, err
	}

	return secretData, nil
}
