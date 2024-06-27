package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/lep13/messaging-notification-service/models"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

// Define getProfileToken as a variable
var getProfileToken = func(profileURL string) (string, error) {
	resp, err := http.Get(profileURL)
	if err != nil {
		log.Printf("Failed to get profile token: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to get profile token, status code: %d", resp.StatusCode)
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tokenResp models.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		log.Printf("Failed to decode token response: %v", err)
		return "", err
	}

	return tokenResp.ProfileToken, nil
}

// ValidateUserProfile sends the token to the mock service and retrieves the user's profile
func ValidateUserProfile(ctx context.Context, sm secretsmanager.SecretManager) (*models.Profile, error) {
	// Fetch the profile URL from secrets manager
	secretName := "notifsecrets"
	secrets, err := sm.GetSecretData(secretName)
	if err != nil {
		log.Printf("Error retrieving secrets: %v", err)
		return nil, err
	}

	profileURL := secrets.ProfileURL

	// Retrieve the profile token from the secure API
	token, err := getProfileToken(profileURL)
	if err != nil {
		log.Printf("Error getting profile token: %v", err)
		return nil, err
	}

	// Prepare the request
	req, err := http.NewRequestWithContext(ctx, "GET", profileURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	// Set the Authorization header with the token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to retrieve user profile: %s", string(body))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var profile models.Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		return nil, err
	}

	// Log the retrieved profile
	log.Printf("Retrieved user profile: %+v\n", profile)
	return &profile, nil
}
