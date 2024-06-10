package main

import (
    // "log"
    "github.com/lep13/messaging-notification-service/kafka"
    // "github.com/lep13/messaging-notification-service/database"
)

func main() {

    brokers := []string{"localhost:9092"}
    topic := "test-topic"

    // Start the Kafka consumer
    kafka.StartConsumer(brokers, topic)


    // // Initialize MongoDB connection
    // database.InitializeMongoDB()

	// err := database.InsertMessage("testUser1", "testUser2", "Hello, this is a test message.", true)
    // if err != nil {
    //     log.Fatalf("Failed to insert test message into MongoDB: %v", err)
    // } else {
    //     log.Println("Test message inserted successfully into MongoDB.")
    // }
}

    // brokerAddress := "EC2_PUBLIC_IP:9092" // Kafka broker on EC2
    // topic := "your-kafka-topic"
    // groupID := "your-consumer-group-id"

    // log.Println("Starting Kafka consumer...")
    // kafka.SetupConsumer(brokerAddress, topic, groupID)

