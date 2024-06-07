package kafka

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumerConfig contains configuration settings for Kafka consumer
type KafkaConsumerConfig struct {
	BrokerAddress string
	Topic         string
	GroupID       string
}

// NewKafkaConsumer creates a new Kafka consumer with the given configuration
func NewKafkaConsumer(config KafkaConsumerConfig) (*kafka.Reader, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{config.BrokerAddress},
		Topic:    config.Topic,
		GroupID:  config.GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	return reader, nil
}

// ConsumeMessages starts consuming messages from Kafka
func ConsumeMessages(ctx context.Context, reader *kafka.Reader) {
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("error while reading message: %v", err)
			continue
		}
		log.Printf("received message at offset %d: %s = %s", msg.Offset, string(msg.Key), string(msg.Value))
	}
}

// SetupConsumer initializes and starts the Kafka consumer
func SetupConsumer(brokerAddress, topic, groupID string) {
	config := KafkaConsumerConfig{
		BrokerAddress: brokerAddress,
		Topic:         topic,
		GroupID:       groupID,
	}

	reader, err := NewKafkaConsumer(config)
	if err != nil {
		log.Fatalf("failed to create Kafka consumer: %v", err)
	}
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ConsumeMessages(ctx, reader)

	// Wait for termination signal to gracefully shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("Shutting down Kafka consumer")
}
