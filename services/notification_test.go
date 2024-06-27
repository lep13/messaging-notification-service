package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lep13/messaging-notification-service/models"
	"github.com/stretchr/testify/assert"
)

// MockSecretManager simulates the behavior of a real secret manager.
type MockSecretManager struct {
	GetSecretDataFunc func(secretName string) (models.SecretData, error)
}

func (m *MockSecretManager) GetSecretData(secretName string) (models.SecretData, error) {
	if m.GetSecretDataFunc != nil {
		return m.GetSecretDataFunc(secretName)
	}
	return models.SecretData{}, errors.New("mock GetSecretData not implemented")
}

func TestNotifyUI(t *testing.T) {
	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{
				NotificationEndpoint: "http://mock-notification-endpoint",
			}, nil
		},
	}

	notification := models.Notification{
		From:    "test-from",
		To:      "test-to",
		Message: "test-message",
	}

	// Mock server to simulate the notification endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var notif models.Notification
		err := json.NewDecoder(r.Body).Decode(&notif)
		assert.NoError(t, err)
		assert.Equal(t, notification, notif)

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Use the mock server URL for the notification endpoint
	mockSM.GetSecretDataFunc = func(secretName string) (models.SecretData, error) {
		return models.SecretData{
			NotificationEndpoint: mockServer.URL,
		}, nil
	}

	// Override the httpClient to avoid actual HTTP requests in tests
	originalHttpClient := httpClient
	httpClient = mockServer.Client()
	defer func() {
		httpClient = originalHttpClient
	}()

	err := NotifyUI(notification, "test-token", mockSM)
	assert.NoError(t, err)
}

func TestNotifyUI_SecretManagerError(t *testing.T) {
	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{}, errors.New("failed to retrieve secrets")
		},
	}

	notification := models.Notification{
		From:    "test-from",
		To:      "test-to",
		Message: "test-message",
	}

	err := NotifyUI(notification, "test-token", mockSM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error retrieving secrets")
}

func TestNotifyUI_HTTPError(t *testing.T) {
	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{
				NotificationEndpoint: "http://mock-notification-endpoint",
			}, nil
		},
	}

	notification := models.Notification{
		From:    "test-from",
		To:      "test-to",
		Message: "test-message",
	}

	// Mock server to simulate the notification endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	// Use the mock server URL for the notification endpoint
	mockSM.GetSecretDataFunc = func(secretName string) (models.SecretData, error) {
		return models.SecretData{
			NotificationEndpoint: mockServer.URL,
		}, nil
	}

	// Override the httpClient to avoid actual HTTP requests in tests
	originalHttpClient := httpClient
	httpClient = mockServer.Client()
	defer func() {
		httpClient = originalHttpClient
	}()

	err := NotifyUI(notification, "test-token", mockSM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification request failed with status")
}

func TestNotifyUI_JSONMarshalError(t *testing.T) {
	// Define a type that will cause json.Marshal to fail
	type UnmarshalableType struct {
		From    string
		To      string
		Message chan int // Channels can't be marshalled to JSON, will cause an error
	}

	// Create an instance of the problematic type
	notification := UnmarshalableType{
		From:    "test-from",
		To:      "test-to",
		Message: make(chan int),
	}

	// Mock server to simulate the notification endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Use the mock server URL for the notification endpoint
	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{
				NotificationEndpoint: mockServer.URL,
			}, nil
		},
	}

	// Override the httpClient to avoid actual HTTP requests in tests
	originalHttpClient := httpClient
	httpClient = mockServer.Client()
	defer func() {
		httpClient = originalHttpClient
	}()

	// Attempt to call NotifyUI with a type that will fail marshalling
	err := NotifyUI(notification, "test-token", mockSM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal notification")
}

// Test case to simulate request creation failure
func TestNotifyUI_RequestCreationError(t *testing.T) {
	notification := models.Notification{
		From:    "test-from",
		To:      "test-to",
		Message: "test-message",
	}

	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{
				NotificationEndpoint: ":::", // Invalid URL to force request creation error
			}, nil
		},
	}

	err := NotifyUI(notification, "test-token", mockSM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}

// Test case to simulate HTTP request failure
func TestNotifyUI_HTTPRequestError(t *testing.T) {
	notification := models.Notification{
		From:    "test-from",
		To:      "test-to",
		Message: "test-message",
	}

	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{
				NotificationEndpoint: "http://invalid-url", // Invalid URL to force HTTP request error
			}, nil
		},
	}

	err := NotifyUI(notification, "test-token", mockSM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send notification")
}