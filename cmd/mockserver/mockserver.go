package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Notification represents the structure of a notification message
type Notification struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}

// NotificationServer holds the state of the mock server
type NotificationServer struct {
	mu            sync.Mutex
	notifications []Notification
	notifyChannel chan Notification
}

// handleNotify handles the incoming notification requests
func (ns *NotificationServer) handleNotify(w http.ResponseWriter, r *http.Request) {
	var notif Notification
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

func main() {
	server := &NotificationServer{
		notifyChannel: make(chan Notification, 100),
	}

	http.HandleFunc("/notify", server.handleNotify)
	http.HandleFunc("/view", server.handleViewNotifications)

	// Start a goroutine to dynamically process notifications
	go server.runDynamicNotifier()

	// Start the server on port 8081
	log.Println("Mock server running on http://localhost:8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Failed to start mock server: %v", err)
	}
}
