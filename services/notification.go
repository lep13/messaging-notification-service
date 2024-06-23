package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lep13/messaging-notification-service/models"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

// NotifyUI sends a notification to the UI
func NotifyUI(notification models.Notification, token string) error {
	// Create a secret manager instance
	secretManager := secretsmanager.NewSecretManager(nil)

	// Fetch the notification endpoint URL from secrets manager
	secretName := "notifsecrets"
	secrets, err := secretManager.GetSecretData(secretName)
	if err != nil {
		log.Printf("Error retrieving secrets: %v", err)
		return fmt.Errorf("error retrieving secrets: %v", err)
	}

	notificationEndpoint := secrets.NotificationEndpoint

	// Marshal the notification to JSON
	payload, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return fmt.Errorf("failed to marshal notification: %v", err)
	}

	req, err := http.NewRequest("POST", notificationEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set authorization token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send notification: %v", err)
		return fmt.Errorf("failed to send notification: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("Notification request failed with status: %v", resp.Status)
		return fmt.Errorf("notification request failed with status: %v", resp.Status)
	}

	log.Printf("Notification sent to UI: From:%s, To:%s, Message:%s", notification.From, notification.To, notification.Message)
	return nil
}
