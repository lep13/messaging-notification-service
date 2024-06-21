package models

// MongoCredentials represents MongoDB credentials structure (not used now)
type MongoCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SecretData represents the structure for all secrets including MongoDB URI and Cognito token
type SecretData struct {
	MongoDBURI           string `json:"mongodbURI"` // MongoDB connection URI
	CognitoToken         string `json:"COGNITO_TOKEN"`  //AWS cognito token
	Region               string `json:"region"`               // AWS region
	UserPoolID           string `json:"userPoolId"`           // User Pool ID
	ProfileURL           string `json:"profileURL"`           // Profile URL
	KafkaBroker          string `json:"kafkaBroker"`          // Kafka Broker IP
	KafkaTopic           string `json:"kafkaTopic"`           // Kafka Topic
	NotificationEndpoint string `json:"notificationEndpoint"` // Notification Endpoint URL
}
