package models

// Notification represents the structure of a notification message
type Notification struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}

// MongoCredentials represents MongoDB credentials structure
type MongoCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
