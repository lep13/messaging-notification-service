package models

// Profile represents the user's profile structure
type Profile struct {
	FirstName string `json:"first_name"`
	UID       string `json:"uid"`
}

// TokenResponse represents the structure for the response from the token API
type TokenResponse struct {
	ProfileToken string `json:"profile_token"`
}

