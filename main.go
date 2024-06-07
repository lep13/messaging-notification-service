package main

import (
	"log"

	// "github.com/segmentio/kafka-go"
	"github.com/lep13/messaging-notification-service/kafka"
)

func main() {
	brokerAddress := "PUBLIC_IP:9092" // Public IP of the EC2 instance and port Kafka is running on
	topic := "your-topic-name"
	groupID := "your-group-id"

	log.Println("Starting Kafka consumer...")
	kafka.SetupConsumer(brokerAddress, topic, groupID)
}
