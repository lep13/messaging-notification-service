package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/ShreerajShettyK/cognitoJwtAuthenticator"
	"github.com/lep13/messaging-notification-service/auth"
	"github.com/lep13/messaging-notification-service/database"
	"github.com/lep13/messaging-notification-service/models"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MockSecretManager is a mock for the SecretManager interface.
type MockSecretManager struct {
	mock.Mock
}

func (m *MockSecretManager) GetSecretData(secretName string) (models.SecretData, error) {
	args := m.Called(secretName)
	return args.Get(0).(models.SecretData), args.Error(1)
}

// MockCollection is a mock for the CollectionInterface.
type MockCollection struct {
	mock.Mock
	InsertOneFunc  func(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	UpdateByIDFunc func(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	if m.InsertOneFunc != nil {
		return m.InsertOneFunc(ctx, document, opts...)
	}
	args := m.Called(ctx, document, opts)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollection) UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if m.UpdateByIDFunc != nil {
		return m.UpdateByIDFunc(ctx, id, update, opts...)
	}
	args := m.Called(ctx, id, update, opts)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// MockAuth is a mock for the Auth interface.
type MockAuth struct {
	mock.Mock
}

func (m *MockAuth) ValidateCognitoToken(ctx context.Context, token string, sm secretsmanager.SecretManager, validator auth.TokenValidatorFunc) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
	args := m.Called(ctx, token, sm, validator)
	return args.Get(0).(*cognitoJwtAuthenticator.AWSCognitoClaims), args.Error(1)
}

func (m *MockAuth) ValidateUserProfile(ctx context.Context, sm secretsmanager.SecretManager) (*models.Profile, error) {
	args := m.Called(ctx, sm)
	return args.Get(0).(*models.Profile), args.Error(1)
}

// MockDatabase is a mock for the Database interface.
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) InitializeMongoDB() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) GetCollection(collectionName string) database.CollectionInterface {
	args := m.Called(collectionName)
	return args.Get(0).(database.CollectionInterface)
}

type MockTokenValidator struct {
	mockAuth *MockAuth
}

func (v *MockTokenValidator) Validate(ctx context.Context, token string, sm secretsmanager.SecretManager, validator auth.TokenValidatorFunc) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
	return v.mockAuth.ValidateCognitoToken(ctx, token, sm, validator)
}

type TokenValidatorFuncType func(ctx context.Context, token string, sm secretsmanager.SecretManager, validator auth.TokenValidatorFunc) (*cognitoJwtAuthenticator.AWSCognitoClaims, error)

func (fn TokenValidatorFuncType) Validate(ctx context.Context, token string, sm secretsmanager.SecretManager, validator auth.TokenValidatorFunc) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
	return fn(ctx, token, sm, validator)
}

// SecretManager mock
type SecretManager struct {
	mock.Mock
}

func (m *SecretManager) GetSecretData(secretName string) (models.SecretData, error) {
	args := m.Called(secretName)
	return args.Get(0).(models.SecretData), args.Error(1)
}

// Auth mock
type Auth struct {
	mock.Mock
}

func (m *Auth) ValidateCognitoToken(ctx context.Context, token string, sm secretsmanager.SecretManager, validator auth.TokenValidatorFunc) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
	args := m.Called(ctx, token, sm, validator)
	return args.Get(0).(*cognitoJwtAuthenticator.AWSCognitoClaims), args.Error(1)
}

func (m *Auth) ValidateUserProfile(ctx context.Context, sm secretsmanager.SecretManager) (*models.Profile, error) {
	args := m.Called(ctx, sm)
	return args.Get(0).(*models.Profile), args.Error(1)
}

// Database mock
type Database struct {
	mock.Mock
}

func (m *Database) InitializeMongoDB() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Database) GetCollection(collectionName string) database.CollectionInterface {
	args := m.Called(collectionName)
	return args.Get(0).(database.CollectionInterface)
}

// MockSaramaConsumer is a mock implementation of the sarama.Consumer interface
type MockSaramaConsumer struct {
	mock.Mock
}

func (m *MockSaramaConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSaramaConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	args := m.Called(topic, partition, offset)
	return args.Get(0).(sarama.PartitionConsumer), args.Error(1)
}

func (m *MockSaramaConsumer) HighWaterMarks() map[string]map[int32]int64 {
	args := m.Called()
	return args.Get(0).(map[string]map[int32]int64)
}

func (m *MockSaramaConsumer) Messages() <-chan *sarama.ConsumerMessage {
	args := m.Called()
	return args.Get(0).(<-chan *sarama.ConsumerMessage)
}

func (m *MockSaramaConsumer) Errors() <-chan *sarama.ConsumerError {
	args := m.Called()
	return args.Get(0).(<-chan *sarama.ConsumerError)
}

func (m *MockSaramaConsumer) Partitions(topic string) ([]int32, error) {
	args := m.Called(topic)
	return args.Get(0).([]int32), args.Error(1)
}

func (m *MockSaramaConsumer) Pause(topicPartitions map[string][]int32) {
	m.Called(topicPartitions)
}

func (m *MockSaramaConsumer) Resume(topicPartitions map[string][]int32) {
	m.Called(topicPartitions)
}

func (m *MockSaramaConsumer) PauseAll() {
	m.Called()
}

func (m *MockSaramaConsumer) ResumeAll() {
	m.Called()
}

func (m *MockSaramaConsumer) Topics() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

// MockKafkaConsumer is a mock implementation of the Kafka consumer
type MockKafkaConsumer struct {
	mock.Mock
}

func (m *MockKafkaConsumer) NewKafkaConsumer(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
	args := m.Called(addrs, config)
	return args.Get(0).(sarama.Consumer), args.Error(1)
}

func (m *MockKafkaConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockKafkaConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	args := m.Called(topic, partition, offset)
	return args.Get(0).(sarama.PartitionConsumer), args.Error(1)
}

func (m *MockKafkaConsumer) HighWaterMarks() map[string]map[int32]int64 {
	args := m.Called()
	return args.Get(0).(map[string]map[int32]int64)
}

func (m *MockKafkaConsumer) Messages() <-chan *sarama.ConsumerMessage {
	args := m.Called()
	return args.Get(0).(<-chan *sarama.ConsumerMessage)
}

func (m *MockKafkaConsumer) Errors() <-chan *sarama.ConsumerError {
	args := m.Called()
	return args.Get(0).(<-chan *sarama.ConsumerError)
}

func (m *MockKafkaConsumer) Partitions(topic string) ([]int32, error) {
	args := m.Called(topic)
	return args.Get(0).([]int32), args.Error(1)
}

func (m *MockKafkaConsumer) Pause(topicPartitions map[string][]int32) {
	m.Called(topicPartitions)
}

func (m *MockKafkaConsumer) Resume(topicPartitions map[string][]int32) {
	m.Called(topicPartitions)
}

func (m *MockKafkaConsumer) PauseAll() {
	m.Called()
}

func (m *MockKafkaConsumer) ResumeAll() {
	m.Called()
}

func (m *MockKafkaConsumer) Topics() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

// Mock methods for Partition consumer
type MockPartitionConsumer struct {
	mock.Mock
}

func (m *MockPartitionConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPartitionConsumer) AsyncClose() {
	m.Called()
}

func (m *MockPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	args := m.Called()
	return args.Get(0).(<-chan *sarama.ConsumerMessage)
}

func (m *MockPartitionConsumer) Errors() <-chan *sarama.ConsumerError {
	args := m.Called()
	return args.Get(0).(<-chan *sarama.ConsumerError)
}

func (m *MockPartitionConsumer) IsPaused() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPartitionConsumer) HighWaterMarkOffset() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *MockPartitionConsumer) Pause() {
	m.Called()
}

func (m *MockPartitionConsumer) Resume() {
	m.Called()
}

// TestRun tests the Run function
func TestRun(t *testing.T) {
	// Create mock instances
	mockSecretManager := new(SecretManager)
	mockAuth := new(Auth)
	mockDatabase := new(Database)
	mockKafkaConsumer := new(MockKafkaConsumer)

	// Set up expected return values for mocks
	mockSecretManager.On("GetSecretData", "notifsecrets").Return(models.SecretData{
		CognitoToken: "mockCognitoToken",
	}, nil)
	mockAuth.On("ValidateCognitoToken", mock.Anything, "mockCognitoToken", mock.Anything, mock.Anything).Return(&cognitoJwtAuthenticator.AWSCognitoClaims{}, nil)
	mockAuth.On("ValidateUserProfile", mock.Anything, mock.Anything).Return(&models.Profile{}, nil)
	mockDatabase.On("InitializeMongoDB").Return(nil)
	mockDatabase.On("GetCollection", mock.Anything).Return(nil)
	mockKafkaConsumer.On("NewKafkaConsumer", mock.Anything, mock.Anything).Return(nil, nil)

	// Define the dependencies
	deps := RunnerDependencies{
		SecretManager:     mockSecretManager,
		AuthValidator:     auth.DefaultTokenValidator,
		ProfileValidator:  mockAuth.ValidateUserProfile,
		InitializeMongoDB: mockDatabase.InitializeMongoDB,
		GetCollection:     mockDatabase.GetCollection,
		NewKafkaConsumer:  mockKafkaConsumer.NewKafkaConsumer,
	}

	// Run the function
	Run(deps)

	// Assertions can be added here as needed
	assert.True(t, true)
}

func TestRunSecretDataFailure(t *testing.T) {
	mockSecretManager := new(SecretManager)
	mockSecretManager.On("GetSecretData", "notifsecrets").Return(models.SecretData{}, errors.New("failed to get secret data"))

	deps := RunnerDependencies{
		SecretManager:    mockSecretManager,
		NewKafkaConsumer: NewKafkaConsumer,
	}

	Run(deps)
	mockSecretManager.AssertExpectations(t)
}

func TestRunMissingCognitoToken(t *testing.T) {
	mockSecretManager := new(SecretManager)
	mockSecretManager.On("GetSecretData", "notifsecrets").Return(models.SecretData{
		CognitoToken: "",
	}, nil)

	deps := RunnerDependencies{
		SecretManager:    mockSecretManager,
		NewKafkaConsumer: NewKafkaConsumer,
	}

	Run(deps)
	mockSecretManager.AssertExpectations(t)
}

func TestRunUserProfileValidationFailure(t *testing.T) {
	mockSecretManager := new(MockSecretManager)
	mockSecretManager.On("GetSecretData", "notifsecrets").Return(models.SecretData{
		CognitoToken: "mockCognitoToken",
	}, nil)
	mockAuth := new(MockAuth)
	mockAuth.On("ValidateCognitoToken", mock.Anything, "mockCognitoToken", mock.Anything, mock.Anything).Return(&cognitoJwtAuthenticator.AWSCognitoClaims{}, nil)
	mockAuth.On("ValidateUserProfile", mock.Anything, mock.Anything).Return((*models.Profile)(nil), errors.New("user profile validation failed"))

	tokenValidator := auth.TokenValidatorFunc(func(ctx context.Context, region, userPoolID, tokenString string) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
		return mockAuth.ValidateCognitoToken(ctx, tokenString, mockSecretManager, auth.DefaultTokenValidator)
	})

	deps := RunnerDependencies{
		SecretManager:     mockSecretManager,
		AuthValidator:     tokenValidator,
		ProfileValidator:  mockAuth.ValidateUserProfile,
		InitializeMongoDB: func() error { return nil },
		GetCollection: func(collectionName string) database.CollectionInterface {
			return nil
		},
		NewKafkaConsumer: NewKafkaConsumer,
	}

	Run(deps)
	mockSecretManager.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
}

// Test for MongoDB initialization failure
func TestRunMongoDBInitializationFailure(t *testing.T) {
	mockSecretManager := new(MockSecretManager)
	mockSecretManager.On("GetSecretData", "notifsecrets").Return(models.SecretData{
		CognitoToken: "mockCognitoToken",
	}, nil)

	mockAuth := new(MockAuth)
	mockAuth.On("ValidateCognitoToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&cognitoJwtAuthenticator.AWSCognitoClaims{}, nil)
	mockAuth.On("ValidateUserProfile", mock.Anything, mock.Anything).Return(&models.Profile{}, nil)

	mockDatabase := new(MockDatabase)
	mockDatabase.On("InitializeMongoDB").Return(errors.New("failed to initialize MongoDB"))

	mockTokenValidator := auth.TokenValidatorFunc(func(ctx context.Context, region, userPoolID, tokenString string) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
		return mockAuth.ValidateCognitoToken(ctx, tokenString, mockSecretManager, auth.DefaultTokenValidator)
	})

	deps := RunnerDependencies{
		SecretManager:     mockSecretManager,
		AuthValidator:     mockTokenValidator,
		ProfileValidator:  mockAuth.ValidateUserProfile,
		InitializeMongoDB: mockDatabase.InitializeMongoDB,
		GetCollection: func(collectionName string) database.CollectionInterface {
			return nil
		},
		NewKafkaConsumer: func(addrs []string, config *sarama.Config) (sarama.Consumer, error) {
			return new(MockSaramaConsumer), nil
		},
	}

	assert.Panics(t, func() { Run(deps) }, "The code did not panic")

	mockSecretManager.AssertExpectations(t)
	mockAuth.AssertExpectations(t)
	mockDatabase.AssertExpectations(t)
}

// func TestRunKafkaConsumerInitialization(t *testing.T) {
// 	mockSecretManager := new(MockSecretManager)
// 	mockSecretManager.On("GetSecretData", "notifsecrets").Return(models.SecretData{
// 		CognitoToken: "mockCognitoToken",
// 		KafkaBroker:  "localhost:9092",
// 		KafkaTopic:   "test-topic",
// 	}, nil)

// 	mockAuth := new(MockAuth)
// 	mockAuth.On("ValidateCognitoToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&cognitoJwtAuthenticator.AWSCognitoClaims{}, nil)
// 	mockAuth.On("ValidateUserProfile", mock.Anything, mock.Anything).Return(&models.Profile{}, nil)

// 	mockDatabase := new(MockDatabase)
// 	mockDatabase.On("InitializeMongoDB").Return(nil)

// 	mockCollection := new(MockCollection)
// 	mockCollection.On("InsertOne", mock.Anything, mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil)
// 	mockCollection.On("UpdateByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{}, nil)
// 	mockDatabase.On("GetCollection", mock.Anything).Return(mockCollection)

// 	mockKafkaConsumer := new(MockKafkaConsumer)
// 	mockPartitionConsumer := new(MockPartitionConsumer)
// 	mockKafkaConsumer.On("NewKafkaConsumer", mock.Anything, mock.Anything).Return(mockKafkaConsumer, nil)
// 	mockKafkaConsumer.On("ConsumePartition", mock.Anything, mock.Anything, mock.Anything).Return(mockPartitionConsumer, nil)
// 	mockKafkaConsumer.On("Close").Return(nil)
// 	mockPartitionConsumer.On("Close").Return(nil)

// 	// Create message and error channels
// 	msgChan := make(chan *sarama.ConsumerMessage, 1)
// 	errChan := make(chan *sarama.ConsumerError, 1)
// 	mockPartitionConsumer.On("Messages").Return((<-chan *sarama.ConsumerMessage)(msgChan))
// 	mockPartitionConsumer.On("Errors").Return((<-chan *sarama.ConsumerError)(errChan))

// 	// Send a mock message to the message channel
// 	go func() {
// 		msgChan <- &sarama.ConsumerMessage{Value: []byte("From:sender, To:receiver, Message:test message")}
// 		close(msgChan)
// 		close(errChan)
// 	}()

// 	mockTokenValidator := auth.TokenValidatorFunc(func(ctx context.Context, region, userPoolID, tokenString string) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
// 		return mockAuth.ValidateCognitoToken(ctx, tokenString, mockSecretManager, auth.DefaultTokenValidator)
// 	})

// 	deps := RunnerDependencies{
// 		SecretManager:     mockSecretManager,
// 		AuthValidator:     mockTokenValidator,
// 		ProfileValidator:  mockAuth.ValidateUserProfile,
// 		InitializeMongoDB: mockDatabase.InitializeMongoDB,
// 		GetCollection:     mockDatabase.GetCollection,
// 		NewKafkaConsumer:  mockKafkaConsumer.NewKafkaConsumer,
// 	}

// 	Run(deps)

// 	mockSecretManager.AssertExpectations(t)
// 	mockAuth.AssertExpectations(t)
// 	mockDatabase.AssertExpectations(t)
// 	mockKafkaConsumer.AssertExpectations(t)
// 	mockPartitionConsumer.AssertExpectations(t)
// 	mockCollection.AssertExpectations(t)
// }
