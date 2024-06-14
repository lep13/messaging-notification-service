package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/lep13/messaging-notification-service/models"
	"github.com/lep13/messaging-notification-service/secrets-manager"
)

// getProfileToken fetches the profile token from the remote API
func getProfileToken(profileURL string) (string, error) {
	resp, err := http.Get(profileURL)
	if err != nil {
		return "", fmt.Errorf("failed to get profile token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get profile token, status code: %d", resp.StatusCode)
	}

	var tokenResp models.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	return tokenResp.ProfileToken, nil
}

// ValidateUserProfile sends the token to the mock service and retrieves the user's profile
func ValidateUserProfile(ctx context.Context) (*models.Profile, error) {
	// Fetch the profile URL from secrets manager
	secretName := "notifsecrets"
	secrets, err := secretsmanager.GetSecretData(secretName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving secrets: %v", err)
	}

	profileURL := secrets.ProfileURL

	// Retrieve the profile token from the secure API
	token, err := getProfileToken(profileURL)
	if err != nil {
		return nil, fmt.Errorf("error getting profile token: %v", err)
	}

	// Prepare the request
	req, err := http.NewRequestWithContext(ctx, "GET", profileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set the Authorization header with the token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve user profile: %s", string(body))
	}

	var profile models.Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	// Log the retrieved profile
	log.Printf("Retrieved user profile: %+v\n", profile)
	return &profile, nil
}
