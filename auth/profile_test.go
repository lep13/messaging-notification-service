package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the getProfileToken function directly
func TestGetProfileToken(t *testing.T) {
	// Test cases for different responses and errors
	tests := []struct {
		name       string
		mockServer *httptest.Server
		wantToken  string
		wantErr    bool
	}{
		{
			name: "Successful Token Fetch",
			mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"profile_token": "mock-token"}`))
			})),
			wantToken: "mock-token",
			wantErr:   false,
		},
		{
			name: "Failed HTTP Request",
			mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})),
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "Unexpected Status Code",
			mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})),
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "Failed Token Decode",
			mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// Intentionally sending invalid JSON
				w.Write([]byte(`{"invalid_json": "mock-token"`))
			})),
			wantToken: "",
			wantErr:   true,
		},
		{
			name:       "Invalid URL",
			mockServer: nil, // No server needed for this case
			wantToken:  "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockServer != nil {
				defer tt.mockServer.Close()
			}
			// Use a valid server URL if available, otherwise an invalid one for the invalid URL case
			url := ""
			if tt.mockServer != nil {
				url = tt.mockServer.URL
			}
			token, err := getProfileToken(url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestValidateUserProfile(t *testing.T) {
	ctx := context.Background()

	// Mock HTTP server to simulate profile URL response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"first_name": "John", "uid": "12345"}`))
	}))
	defer mockServer.Close()

	// Using the shared mock secrets manager and setting the ProfileURL to the mock server's URL
	mockSM := &MockSecretsManager{
		ProfileURL: mockServer.URL,
	}

	// Temporarily override getProfileToken to return the mock token
	originalGetProfileToken := getProfileToken
	getProfileToken = func(profileURL string) (string, error) {
		if profileURL == "" {
			return "", errors.New("invalid URL")
		}
		return "mock-token", nil
	}
	defer func() { getProfileToken = originalGetProfileToken }() // Restore original function after test

	profile, err := ValidateUserProfile(ctx, mockSM)

	assert.Nil(t, err)
	assert.NotNil(t, profile)

	// Additional test cases to cover more error scenarios
	t.Run("Error Retrieving Profile Token", func(t *testing.T) {
		// Override getProfileToken to simulate an error
		getProfileToken = func(profileURL string) (string, error) {
			return "", errors.New("failed to fetch token")
		}

		profile, err := ValidateUserProfile(ctx, mockSM)
		assert.Error(t, err)
		assert.Nil(t, profile)
	})

	t.Run("Error Creating HTTP Request", func(t *testing.T) {
		// Use an invalid URL to simulate request creation failure
		mockSM.ProfileURL = "://invalid-url"
		profile, err := ValidateUserProfile(ctx, mockSM)
		assert.Error(t, err)
		assert.Nil(t, profile)
	})

	t.Run("HTTP Request Fails", func(t *testing.T) {
		// Use a valid URL but simulate a server not responding
		mockServer.Close() // Close the mock server to simulate request failure
		mockSM.ProfileURL = mockServer.URL
		profile, err := ValidateUserProfile(ctx, mockSM)
		assert.Error(t, err)
		assert.Nil(t, profile)
	})

	t.Run("Error Reading Response", func(t *testing.T) {
		// Mock server that closes the connection before sending the response body
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			// w.Close()
		}))
		defer mockServer.Close()

		mockSM.ProfileURL = mockServer.URL
		profile, err := ValidateUserProfile(ctx, mockSM)
		assert.Error(t, err)
		assert.Nil(t, profile)
	})

	t.Run("Error Unmarshalling Response", func(t *testing.T) {
		// Mock server that sends invalid JSON
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"invalid_json"`))
		}))
		defer mockServer.Close()

		mockSM.ProfileURL = mockServer.URL
		profile, err := ValidateUserProfile(ctx, mockSM)
		assert.Error(t, err)
		assert.Nil(t, profile)
	})
}
