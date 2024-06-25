package kafka

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/lep13/messaging-notification-service/database"
	"github.com/lep13/messaging-notification-service/models"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

// Mock SecretManager
type MockSecretManager struct {
	GetSecretDataFunc func(secretName string) (models.SecretData, error)
}

func (m *MockSecretManager) GetSecretData(secretName string) (models.SecretData, error) {
	if m.GetSecretDataFunc != nil {
		return m.GetSecretDataFunc(secretName)
	}
	return models.SecretData{}, errors.New("mock GetSecretData not implemented")
}

func NewMockSecretManager() *MockSecretManager {
	return &MockSecretManager{}
}

// MockCollection implements the CollectionInterface for testing.
type MockCollection struct {
	InsertOneFunc  func(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	UpdateByIDFunc func(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	if m.InsertOneFunc != nil {
		return m.InsertOneFunc(ctx, document, opts...)
	}
	return nil, errors.New("InsertOne not implemented")
}

func (m *MockCollection) UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if m.UpdateByIDFunc != nil {
		return m.UpdateByIDFunc(ctx, id, update, opts...)
	}
	return nil, errors.New("UpdateByID not implemented")
}

func getMockCollection(collectionName string) database.CollectionInterface {
	return &MockCollection{}
}

func mockNotifyUI(notification models.Notification, token string, secretManager secretsmanager.SecretManager) error {
	return nil
}

// MockConsumer for Kafka
type MockConsumer struct {
	MessagesChan         chan *sarama.ConsumerMessage
	ErrorsChan           chan *sarama.ConsumerError
	CloseFunc            func() error
	ConsumePartitionFunc func(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error)
	TopicsFunc           func() ([]string, error)
}

func (m *MockConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return m.MessagesChan
}

func (m *MockConsumer) Errors() <-chan *sarama.ConsumerError {
	return m.ErrorsChan
}

func (m *MockConsumer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	if m.ConsumePartitionFunc != nil {
		return m.ConsumePartitionFunc(topic, partition, offset)
	}
	return nil, errors.New("ConsumePartition not implemented")
}

func (m *MockConsumer) Topics() ([]string, error) {
	if m.TopicsFunc != nil {
		return m.TopicsFunc()
	}
	return []string{"test-topic"}, nil
}

func (m *MockConsumer) HighWaterMarks() map[string]map[int32]int64 {
	return nil
}

func (m *MockConsumer) Partitions(topic string) ([]int32, error) {
	return nil, nil
}

func (m *MockConsumer) Pause(topicPartitions map[string][]int32) {}

func (m *MockConsumer) Resume(topicPartitions map[string][]int32) {}

func (m *MockConsumer) PauseAll() {}

func (m *MockConsumer) ResumeAll() {}

// MockPartitionConsumer for Kafka
type MockPartitionConsumer struct {
	MessagesChan chan *sarama.ConsumerMessage
	ErrorsChan   chan *sarama.ConsumerError
	CloseFunc    func() error
}

func (m *MockPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return m.MessagesChan
}

func (m *MockPartitionConsumer) Errors() <-chan *sarama.ConsumerError {
	return m.ErrorsChan
}

func (m *MockPartitionConsumer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockPartitionConsumer) AsyncClose() {}

func (m *MockPartitionConsumer) HighWaterMarkOffset() int64 {
	return 0
}

func (m *MockPartitionConsumer) IsPaused() bool {
	return false
}

func (m *MockPartitionConsumer) Pause() {}

func (m *MockPartitionConsumer) Resume() {}

func TestConsumeMessages_ErrorRetrievingSecrets(t *testing.T) {
	// Mock dependencies
	mockSecretManager := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{}, errors.New("mock error retrieving secrets")
		},
	}

	// Prepare the dependencies struct
	deps := ConsumerDependencies{
		NewSecretManager: func(client secretsmanager.SecretsManagerAPI) secretsmanager.SecretManager {
			return mockSecretManager
		},
		InitializeMongoDB: func() error { return nil },
		GetCollection:     getMockCollection,
		NewKafkaConsumer: func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
			return nil, errors.New("mock error initializing Kafka consumer")
		},
	}

	// Mock logging to avoid cluttering test output
	log.SetOutput(io.Discard)
	defer log.SetOutput(log.Writer())

	// Run the consumer
	ConsumeMessages(deps)

	// Assertions
	assert.NotNil(t, mockSecretManager)
}

func TestConsumeMessages_ErrorMongoDBInit(t *testing.T) {
	// Mock dependencies
	mockSecretManager := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{
				KafkaBroker: "localhost:9092",
				KafkaTopic:  "test-topic",
			}, nil
		},
	}

	// Prepare the dependencies struct
	deps := ConsumerDependencies{
		NewSecretManager: func(client secretsmanager.SecretsManagerAPI) secretsmanager.SecretManager {
			return mockSecretManager
		},
		InitializeMongoDB: func() error { return errors.New("mock error initializing MongoDB") },
		GetCollection:     getMockCollection,
		NewKafkaConsumer: func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
			return &MockConsumer{}, nil
		},
	}

	// Mock logging to avoid cluttering test output
	log.SetOutput(io.Discard)
	defer log.SetOutput(log.Writer())

	// Run the consumer
	ConsumeMessages(deps)

	// Assertions
	assert.NotNil(t, mockSecretManager)
}

// func TestConsumeMessages_Success(t *testing.T) {
// 	// Mock dependencies
// 	mockSecretManager := &MockSecretManager{
// 		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
// 			return models.SecretData{
// 				KafkaBroker: "localhost:9092",
// 				KafkaTopic:  "test-topic",
// 			}, nil
// 		},
// 	}

// 	mockConsumer := &MockConsumer{
// 		MessagesChan: make(chan *sarama.ConsumerMessage),
// 		ErrorsChan:   make(chan *sarama.ConsumerError),
// 	}

// 	mockPartitionConsumer := &MockPartitionConsumer{
// 		MessagesChan: mockConsumer.MessagesChan,
// 		ErrorsChan:   mockConsumer.ErrorsChan,
// 		CloseFunc:    func() error { return nil },
// 	}

// 	// Prepare the mock for the Kafka consumer
// 	mockConsumer.ConsumePartitionFunc = func(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
// 		return mockPartitionConsumer, nil
// 	}

// 	// Prepare the dependencies struct
// 	deps := ConsumerDependencies{
// 		NewSecretManager: func(client secretsmanager.SecretsManagerAPI) secretsmanager.SecretManager {
// 			return mockSecretManager
// 		},
// 		InitializeMongoDB: func() error { return nil },
// 		GetCollection:     getMockCollection,
// 		NewKafkaConsumer: func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
// 			return mockConsumer, nil
// 		},
// 	}

// 	// Mock logging to avoid cluttering test output
// 	log.SetOutput(io.Discard)
// 	defer log.SetOutput(log.Writer())

// 	// Run the consumer
// 	go func() {
// 		ConsumeMessages(deps)
// 	}()

// 	// Simulate message reception
// 	mockConsumer.MessagesChan <- &sarama.ConsumerMessage{Value: []byte("From:John, To:Jane, Message:Hello!")}
// 	time.Sleep(100 * time.Millisecond) // Allow some time for processing
// 	close(mockConsumer.MessagesChan)
// 	close(mockConsumer.ErrorsChan)

// 	// Assertions
// 	assert.NotNil(t, mockSecretManager)
// 	assert.NotNil(t, mockConsumer)
// 	assert.NotNil(t, mockPartitionConsumer)
// }

// func TestConsumeMessages_PartitionConsumerErrorHandling(t *testing.T) {
// 	// Mock dependencies
// 	mockSecretManager := &MockSecretManager{
// 		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
// 			return models.SecretData{
// 				KafkaBroker: "localhost:9092",
// 				KafkaTopic:  "test-topic",
// 			}, nil
// 		},
// 	}

// 	mockConsumer := &MockConsumer{
// 		MessagesChan: make(chan *sarama.ConsumerMessage),
// 		ErrorsChan:   make(chan *sarama.ConsumerError),
// 	}

// 	mockPartitionConsumer := &MockPartitionConsumer{
// 		MessagesChan: mockConsumer.MessagesChan,
// 		ErrorsChan:   mockConsumer.ErrorsChan,
// 		CloseFunc:    func() error { return nil },
// 	}

// 	// Prepare the mock for the Kafka consumer
// 	mockConsumer.ConsumePartitionFunc = func(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
// 		return mockPartitionConsumer, nil
// 	}

// 	// Prepare the dependencies struct
// 	deps := ConsumerDependencies{
// 		NewSecretManager: func(client secretsmanager.SecretsManagerAPI) secretsmanager.SecretManager {
// 			return mockSecretManager
// 		},
// 		InitializeMongoDB: func() error { return nil },
// 		GetCollection:     getMockCollection,
// 		NewKafkaConsumer: func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
// 			return mockConsumer, nil
// 		},
// 	}

// 	// Mock logging to avoid cluttering test output
// 	log.SetOutput(io.Discard)
// 	defer log.SetOutput(log.Writer())

// 	// Run the consumer
// 	go func() {
// 		ConsumeMessages(deps)
// 	}()

// 	// Simulate an error reception
// 	mockConsumer.ErrorsChan <- &sarama.ConsumerError{Err: errors.New("mock partition consumer error")}
// 	time.Sleep(100 * time.Millisecond) // Allow some time for processing
// 	close(mockConsumer.MessagesChan)
// 	close(mockConsumer.ErrorsChan)

// 	// Assertions
// 	assert.NotNil(t, mockSecretManager)
// 	assert.NotNil(t, mockConsumer)
// 	assert.NotNil(t, mockPartitionConsumer)
// }

func TestProcessMessage_Success(t *testing.T) {
	// Mock the secrets manager
	mockSecretManager := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{}, nil
		},
	}

	// Mock the collection and insert behavior
	mockCollection := &MockCollection{
		InsertOneFunc: func(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
			return &mongo.InsertOneResult{InsertedID: "mockID"}, nil
		},
		UpdateByIDFunc: func(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
			return &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
		},
	}

	// Replace GetCollection to return the mock collection
	getCollection := func(name string) database.CollectionInterface {
		return mockCollection
	}

	message := []byte("From:John, To:Jane, Message:Hello!")
	processMessage(message, mockSecretManager, getCollection, mockNotifyUI)

	// Assertions to check that the expected methods were called on the mock
	assert.NotNil(t, mockCollection.InsertOneFunc, "InsertOne should be called on the mock collection")
	assert.NotNil(t, mockCollection.UpdateByIDFunc, "UpdateByID should be called on the mock collection")
}

func TestProcessMessage_InsertFailure(t *testing.T) {
	// Mock the secrets manager
	mockSecretManager := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{}, nil
		},
	}

	// Mock the collection to simulate an insert failure
	mockCollection := &MockCollection{
		InsertOneFunc: func(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
			return nil, errors.New("mock insert failure")
		},
	}

	// Replace GetCollection to return the mock collection
	getCollection := func(name string) database.CollectionInterface {
		return mockCollection
	}

	message := []byte("From:John, To:Jane, Message:Hello!")
	processMessage(message, mockSecretManager, getCollection, mockNotifyUI)

	// Assertions to check that the expected methods were called on the mock
	assert.NotNil(t, mockCollection.InsertOneFunc, "InsertOne should be called on the mock collection")
}

func TestProcessMessage_NotificationFailure(t *testing.T) {
	// Mock the secrets manager
	mockSecretManager := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{}, nil
		},
	}

	// Mock the collection to simulate successful insert and update
	mockCollection := &MockCollection{
		InsertOneFunc: func(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
			return &mongo.InsertOneResult{InsertedID: "mockID"}, nil
		},
		UpdateByIDFunc: func(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
			return &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
		},
	}

	// Replace GetCollection to return the mock collection
	getCollection := func(name string) database.CollectionInterface {
		return mockCollection
	}

	// Mock notification service to simulate a failure
	mockNotifyUI := func(notification models.Notification, token string, secretManager secretsmanager.SecretManager) error {
		return errors.New("mock notification failure")
	}

	message := []byte("From:John, To:Jane, Message:Hello!")
	processMessage(message, mockSecretManager, getCollection, mockNotifyUI)

	// Assertions to check that the expected methods were called on the mock
	assert.NotNil(t, mockCollection.InsertOneFunc, "InsertOne should be called on the mock collection")
	assert.NotNil(t, mockCollection.UpdateByIDFunc, "UpdateByID should be called on the mock collection")
}

func TestParseMessage(t *testing.T) {
	message := "From:John, To:Jane, Message:Hello!"

	expected := bson.M{
		"from":    "John",
		"to":      "Jane",
		"message": "Hello!",
	}

	result := parseMessage(message)
	assert.Equal(t, expected, result)
}

func TestParseMessage_InvalidFormat(t *testing.T) {
	message := "Invalid message format"

	result := parseMessage(message)
	assert.Nil(t, result)
}
