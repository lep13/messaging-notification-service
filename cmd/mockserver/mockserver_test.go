package main

import (
	"bytes"
	"encoding/json"
	"io"
	// "net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lep13/messaging-notification-service/models"
	"github.com/stretchr/testify/assert"
)

// Setup a new NotificationServer for each test
func setupServer() *NotificationServer {
	return &NotificationServer{
		notifyChannel: make(chan models.Notification, 100),
		notifications: []models.Notification{},
	}
}

// Test handleNotify endpoint
func TestHandleNotify(t *testing.T) {
	server := setupServer()
	reqBody := models.Notification{From: "test-from", To: "test-to", Message: "test-message"}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/notify", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(server.handleNotify)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Notification received successfully")

	// Check if the notification was added to the server state
	server.mu.Lock()
	defer server.mu.Unlock()
	assert.Equal(t, 1, len(server.notifications))
	assert.Equal(t, reqBody, server.notifications[0])
}

// Test handleNotify with invalid JSON
func TestHandleNotify_InvalidJSON(t *testing.T) {
	server := setupServer()

	// Simulate an invalid JSON request
	invalidJSON := "{invalid-json"
	req, err := http.NewRequest("POST", "/notify", bytes.NewReader([]byte(invalidJSON)))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(server.handleNotify)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Bad request")
}

// Test handleViewNotifications endpoint
func TestHandleViewNotifications(t *testing.T) {
	server := setupServer()

	// Add a notification to the server state
	server.mu.Lock()
	server.notifications = append(server.notifications, models.Notification{From: "test-from", To: "test-to", Message: "test-message"})
	server.mu.Unlock()

	req, err := http.NewRequest("GET", "/view", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(server.handleViewNotifications)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "From: test-from, To: test-to, Message: test-message")

	// Check if the notifications list was cleared
	server.mu.Lock()
	defer server.mu.Unlock()
	assert.Equal(t, 0, len(server.notifications))
}

// Test the dynamic processing of notifications
func TestRunDynamicNotifier(t *testing.T) {
	server := setupServer()

	// Simulate sending notifications to the notifyChannel
	go server.runDynamicNotifier()

	server.notifyChannel <- models.Notification{From: "test-from", To: "test-to", Message: "test-message"}

	// Allow some time for processing (simulate dynamic notification processing)
	time.Sleep(2 * time.Second)

	// Check the logs for processed notifications (no direct way to capture logs, so we'll focus on not crashing)
}

// Test server startup and endpoint routing
func TestServerRouting(t *testing.T) {
	server := setupServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/notify", server.handleNotify)
	mux.HandleFunc("/view", server.handleViewNotifications)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	// Test /notify endpoint
	notifyURL := testServer.URL + "/notify"
	reqBody := models.Notification{From: "test-from", To: "test-to", Message: "test-message"}
	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(notifyURL, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(responseBody), "Notification received successfully")

	// Test /view endpoint
	viewURL := testServer.URL + "/view"
	resp, err = http.Get(viewURL)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()
	responseBody, _ = io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(responseBody), "From: test-from, To: test-to, Message: test-message")
}

// Test server initialization and startup
func TestStartServer(t *testing.T) {
	// Run the server in a separate goroutine to simulate real server startup
	go func() {
		err := StartServer()
		assert.NoError(t, err)
	}()

	// Allow some time for the server to start
	time.Sleep(1 * time.Second)

	// Test if server is running by sending a request
	resp, err := http.Get("http://localhost:8081/view")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// // // Test server startup failure
// func TestStartServer_Failure(t *testing.T) {
// 	// Bind to port 8081 to simulate a conflict
// 	listener, err := net.Listen("tcp", ":8081")
// 	if err != nil {
// 		t.Fatalf("Failed to bind to port 8081: %v", err)
// 	}
// 	defer listener.Close()

// 	// Try to start the server again, which should fail
// 	err = StartServer()
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "address already in use")
// }
