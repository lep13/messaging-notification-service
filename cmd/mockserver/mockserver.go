package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/lep13/messaging-notification-service/models"
)

// NotificationServer holds the state of the mock server
type NotificationServer struct {
	mu            sync.Mutex
	notifications []models.Notification
	notifyChannel chan models.Notification
}

// handleNotify handles the incoming notification requests
func (ns *NotificationServer) handleNotify(w http.ResponseWriter, r *http.Request) {
	var notif models.Notification
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	ns.mu.Lock()
	ns.notifications = append(ns.notifications, notif)
	ns.mu.Unlock()

	ns.notifyChannel <- notif
	fmt.Fprintf(w, "Notification received successfully")
}

// handleViewNotifications dynamically displays notifications with a delay
func (ns *NotificationServer) handleViewNotifications(w http.ResponseWriter, r *http.Request) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if len(ns.notifications) == 0 {
		fmt.Fprintf(w, "No notifications to display\n")
		return
	}

	for _, notif := range ns.notifications {
		fmt.Fprintf(w, "From: %s, To: %s, Message: %s\n", notif.From, notif.To, notif.Message)
		time.Sleep(5 * time.Second) // Simulate delay between notifications
	}
	ns.notifications = nil
}

// runDynamicNotifier listens to the notifyChannel and dynamically processes notifications
func (ns *NotificationServer) runDynamicNotifier() {
	for {
		select {
		case notif := <-ns.notifyChannel:
			log.Printf("Processing notification: From: %s, To: %s, Message: %s\n", notif.From, notif.To, notif.Message)
			time.Sleep(5 * time.Second) // Simulate processing time
		}
	}
}

// StartServer starts the HTTP server and listens on the specified port.
func StartServer() error {
	server := &NotificationServer{
		notifyChannel: make(chan models.Notification, 100),
	}

	http.HandleFunc("/notify", server.handleNotify)
	http.HandleFunc("/view", server.handleViewNotifications)

	// Start a goroutine to dynamically process notifications
	go server.runDynamicNotifier()

	// Start the server on port 8081
	log.Println("Mock server running on http://localhost:8081")
	return http.ListenAndServe(":8081", nil)
}

func main() {
	if err := StartServer(); err != nil {
		log.Printf("Failed to start mock server: %v", err)
	}
}
