package main

import (
    "log"
    "github.com/lep13/messaging-notification-service/kafka"
    "github.com/lep13/messaging-notification-service/database"
)

func main() {
    // Initialize MongoDB connection
    database.InitializeMongoDB()

    // Replace these values with the actual values from your setup
    brokerAddress := "EC2_PUBLIC_IP:9092" // Kafka broker on EC2
    topic := "your-kafka-topic"
    groupID := "your-consumer-group-id"

    log.Println("Starting Kafka consumer...")
    kafka.SetupConsumer(brokerAddress, topic, groupID)
}
