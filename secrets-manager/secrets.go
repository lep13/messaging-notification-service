package secretsmanager

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// MongoCredentials represents MongoDB credentials structure
type MongoCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetMongoCredentials fetches the MongoDB credentials from AWS Secrets Manager
func GetMongoCredentials(secretName string) (*MongoCredentials, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	svc := secretsmanager.New(sess, aws.NewConfig().WithRegion("us-east-1"))

	result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return nil, err
	}

	var credentials MongoCredentials
	err = json.Unmarshal([]byte(*result.SecretString), &credentials)
	if err != nil {
		return nil, err
	}

	return &credentials, nil
}
