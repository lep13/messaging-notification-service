package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/lep13/messaging-notification-service/auth"
	"github.com/lep13/messaging-notification-service/database"
	"github.com/segmentio/kafka-go"
)

// // Message represents the structure of the Kafka message
// type Message struct {
// 	FromUser string `json:"from_user"`
// 	ToUser   string `json:"to_user"`
// 	Content  string `json:"content"`
// }

// func ConsumeMessages(ctx context.Context, reader *kafka.Reader) {
// 	for {
// 		msg, err := reader.ReadMessage(ctx)
// 		if err != nil {
// 			log.Printf("error while reading message: %v", err)
// 			continue
// 		}

// 		var message Message
// 		err = json.Unmarshal(msg.Value, &message)
// 		if err != nil {
// 			log.Printf("error unmarshalling message: %v", err)
// 			continue
// 		}

// 		log.Printf("received message: %+v", message)

// 		// Validate "To" user
// 		isValid, err := auth.ValidateUser(message.ToUser)
// 		if err != nil || !isValid {
// 			log.Printf("error validating user: %v", err)
// 			continue
// 		}

// 		// Store the message in MongoDB
// 		err = database.InsertMessage(message.FromUser, message.ToUser, message.Content, false)
// 		if err != nil {
// 			log.Printf("error inserting message into MongoDB: %v", err)
// 		}
// 	}
// }

import (
    "fmt"
    "log"
    "os"
    "os/signal"

    "github.com/Shopify/sarama"
)

func StartConsumer(brokers []string, topic string) {
    // Set up the Sarama configuration
    config := sarama.NewConfig()
    config.Consumer.Return.Errors = true

    // Create a new consumer
    consumer, err := sarama.NewConsumer(brokers, config)
    if err != nil {
        log.Fatalf("Failed to start consumer: %v", err)
    }
    defer consumer.Close()

    // Get the list of partitions
    partitions, err := consumer.Partitions(topic)
    if err != nil {
        log.Fatalf("Failed to get the list of partitions: %v", err)
    }

    // Handle interrupt signals to allow graceful shutdown
    sigchan := make(chan os.Signal, 1)
    signal.Notify(sigchan, os.Interrupt)

    // Consume messages from each partition
    for _, partition := range partitions {
        partitionConsumer, err := consumer.ConsumePartition(topic, partition, sarama.OffsetOldest)
        if err != nil {
            log.Fatalf("Failed to consume partition %d: %v", partition, err)
        }
        defer partitionConsumer.Close()

        // Start a goroutine for each partition consumer
        go func(pc sarama.PartitionConsumer) {
            for {
                select {
                case msg := <-pc.Messages():
                    fmt.Printf("Consumed message offset %d: %s\n", msg.Offset, string(msg.Value))
                case err := <-pc.Errors():
                    fmt.Printf("Error consuming message: %v\n", err)
                case <-sigchan:
                    fmt.Println("Interrupt is detected. Closing consumer...")
                    return
                }
            }
        }(partitionConsumer)
    }

    // Block until an interrupt signal is received
    <-sigchan
    fmt.Println("Consumer stopped")
}
