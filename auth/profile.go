package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

// Profile represents the user's profile structure
type Profile struct {
	FirstName string `json:"first_name"`
	UID       string `json:"uid"`
}

// TokenResponse represents the structure for the response from the token API
type TokenResponse struct {
	ProfileToken string `json:"profile_token"`
}

// getProfileToken fetches the profile token from the remote API
func getProfileToken() (string, error) {
	secretName := "mongodbcreds"
	secrets, err := secretsmanager.GetSecretData(secretName)
	if err != nil {
		return "", fmt.Errorf("error retrieving secrets: %v", err)
	}

	profileTokenURL := secrets.ProfileTokenURL

	resp, err := http.Get(profileTokenURL)
	if err != nil {
		return "", fmt.Errorf("failed to get profile token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get profile token, status code: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	return tokenResp.ProfileToken, nil
}

// ValidateUserProfile sends the token to the mock service and retrieves the user's profile
func ValidateUserProfile(ctx context.Context) (*Profile, error) {
	// Retrieve the profile token from the secure API
	token, err := getProfileToken()
	if err != nil {
		return nil, fmt.Errorf("error getting profile token: %v", err)
	}

	// Fetch the profile URL from secrets manager
	secretName := "mongodbcreds"
	secrets, err := secretsmanager.GetSecretData(secretName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving secrets: %v", err)
	}

	profileURL := secrets.ProfileURL

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

	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	// Log the retrieved profile
	log.Printf("Retrieved user profile: %+v\n", profile)
	return &profile, nil
}
