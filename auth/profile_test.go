package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lep13/messaging-notification-service/models"
	"github.com/stretchr/testify/assert"
)

type MockSecretsManagerWithError struct {
	MockSecretsManager
	err error
}

func (m *MockSecretsManagerWithError) GetSecretData(secretName string) (models.SecretData, error) {
	return models.SecretData{}, m.err
}

func TestGetProfileToken(t *testing.T) {
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
				w.Write([]byte(`{"invalid_json": "mock-token"`))
			})),
			wantToken: "",
			wantErr:   true,
		},
		{
			name:       "Invalid URL",
			mockServer: nil,
			wantToken:  "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockServer != nil {
				defer tt.mockServer.Close()
			}
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

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"first_name": "John", "uid": "12345"}`))
	}))
	defer mockServer.Close()

	mockSM := &MockSecretsManager{
		ProfileURL: mockServer.URL,
	}

	originalGetProfileToken := getProfileToken
	getProfileToken = func(profileURL string) (string, error) {
		if profileURL == "" {
			return "", errors.New("invalid URL")
		}
		return "mock-token", nil
	}
	defer func() { getProfileToken = originalGetProfileToken }()

	profile, err := ValidateUserProfile(ctx, mockSM)

	assert.Nil(t, err)
	assert.NotNil(t, profile)

	t.Run("Error Retrieving Profile Token", func(t *testing.T) {
		getProfileToken = func(profileURL string) (string, error) {
			return "", errors.New("failed to fetch token")
		}

		profile, err := ValidateUserProfile(ctx, mockSM)
		assert.Error(t, err)
		assert.Nil(t, profile)
	})

	t.Run("Error Creating HTTP Request", func(t *testing.T) {
		mockSM.ProfileURL = "://invalid-url"
		profile, err := ValidateUserProfile(ctx, mockSM)
		assert.Error(t, err)
		assert.Nil(t, profile)
	})

	t.Run("HTTP Request Fails", func(t *testing.T) {
		mockServer.Close()
		mockSM.ProfileURL = mockServer.URL
		profile, err := ValidateUserProfile(ctx, mockSM)
		assert.Error(t, err)
		assert.Nil(t, profile)
	})

	t.Run("Error Reading Response", func(t *testing.T) {
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

func TestValidateUserProfile_ErrorFetchingSecrets(t *testing.T) {
	ctx := context.Background()

	mockSM := &MockSecretsManagerWithError{
		err: errors.New("error fetching secrets"),
	}

	profile, err := ValidateUserProfile(ctx, mockSM)

	assert.NotNil(t, err)
	assert.Nil(t, profile)
}
