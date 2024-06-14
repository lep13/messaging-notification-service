package models

// MongoCredentials represents MongoDB credentials structure
type MongoCredentials struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	MongoDBURI string `json:"mongodbURI"` 
}

// SecretData represents the structure for all secrets including MongoDB credentials and Cognito token
type SecretData struct {
	MongoCredentials    MongoCredentials `json:"mongodbcreds"`
	CognitoToken        string           `json:"COGNITO_TOKEN"`
	Region              string           `json:"region"`           // Added field for AWS region
	UserPoolID          string           `json:"userPoolId"`       // Added field for User Pool ID
	ProfileURL          string           `json:"profileURL"`       // Added field for Profile URL
	KafkaBroker         string           `json:"kafkaBroker"`      // Added field for Kafka Broker IP
	KafkaTopic          string           `json:"kafkaTopic"`       // Added field for Kafka Topic
	NotificationEndpoint string          `json:"notificationEndpoint"` // Added field for Notification Endpoint URL
}
