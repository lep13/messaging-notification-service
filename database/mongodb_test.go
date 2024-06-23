package database

import (
	"context"
	"errors"
	"testing"

	"github.com/lep13/messaging-notification-service/models"
	secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MockSecretManager simulates the behavior of a real secret manager.
type MockSecretManager struct {
	GetSecretDataFunc func(secretName string) (models.SecretData, error)
}

func (m *MockSecretManager) GetSecretData(secretName string) (models.SecretData, error) {
	if m.GetSecretDataFunc != nil {
		return m.GetSecretDataFunc(secretName)
	}
	return models.SecretData{}, errors.New("mock GetSecretData not implemented")
}

// MockMongoClient simulates the MongoDB client behavior.
type MockMongoClient struct {
	PingFunc     func(ctx context.Context, rp *readpref.ReadPref) error
	DatabaseFunc func(name string, opts ...*options.DatabaseOptions) *mongo.Database
}

func (m *MockMongoClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx, rp)
	}
	return nil
}

func (m *MockMongoClient) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
	if m.DatabaseFunc != nil {
		return m.DatabaseFunc(name, opts...)
	}
	return &mongo.Database{} // Return a basic mongo.Database object
}

// MockMongoDatabase simulates the behavior of a MongoDB database.
type MockMongoDatabase struct {
	CollectionFunc func(name string, opts ...*options.CollectionOptions) *mongo.Collection
}

func (m *MockMongoDatabase) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	if m.CollectionFunc != nil {
		return m.CollectionFunc(name, opts...)
	}
	return &mongo.Collection{}
}

// Wrapper for *mongo.Database to satisfy the required interface while using mock implementations.
type MongoDatabaseWrapper struct {
	*MockMongoDatabase
}

func (w *MongoDatabaseWrapper) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return w.MockMongoDatabase.Collection(name, opts...)
}

func TestInitializeMongoDB_Success(t *testing.T) {
	// Mock SecretManager
	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{
				MongoDBURI: "mongodb://mock-uri",
			}, nil
		},
	}

	// Mock Mongo connect function
	originalMongoConnectFunc := DefaultMongoConnectFunc
	defer func() { DefaultMongoConnectFunc = originalMongoConnectFunc }() // Restore original after test
	DefaultMongoConnectFunc = func(ctx context.Context, uri string) (MongoClientInterface, error) {
		return &MockMongoClient{}, nil // Mock success
	}

	// Override SecretManagerFunc to return the mock SecretManager
	originalSecretManagerFunc := SecretManagerFunc
	defer func() { SecretManagerFunc = originalSecretManagerFunc }()
	SecretManagerFunc = func() secretsmanager.SecretManager {
		return mockSM
	}

	err := InitializeMongoDB()
	assert.NoError(t, err)
	assert.NotNil(t, MongoClient)
}

func TestInitializeMongoDB_SecretManagerError(t *testing.T) {
	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{}, errors.New("failed to retrieve secrets")
		},
	}

	// Override SecretManagerFunc to return the mock SecretManager
	originalSecretManagerFunc := SecretManagerFunc
	defer func() { SecretManagerFunc = originalSecretManagerFunc }()
	SecretManagerFunc = func() secretsmanager.SecretManager {
		return mockSM
	}

	err := InitializeMongoDB()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve secrets")
}

func TestInitializeMongoDB_MongoConnectError(t *testing.T) {
	// Mock SecretManager
	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{
				MongoDBURI: "mongodb://mock-uri",
			}, nil
		},
	}

	// Mock Mongo connect function that returns an error
	originalMongoConnectFunc := DefaultMongoConnectFunc
	defer func() { DefaultMongoConnectFunc = originalMongoConnectFunc }() // Restore original after test
	DefaultMongoConnectFunc = func(ctx context.Context, uri string) (MongoClientInterface, error) {
		return nil, errors.New("failed to connect to MongoDB")
	}

	// Override SecretManagerFunc to return the mock SecretManager
	originalSecretManagerFunc := SecretManagerFunc
	defer func() { SecretManagerFunc = originalSecretManagerFunc }()
	SecretManagerFunc = func() secretsmanager.SecretManager {
		return mockSM
	}

	err := InitializeMongoDB()
	assert.Error(t, err)
	// Adjust to the exact error message produced in the InitializeMongoDB function
	assert.Contains(t, err.Error(), "failed to connect to MongoDB")
}

func TestInitializeMongoDB_MongoPingError(t *testing.T) {
	// Mock SecretManager
	mockSM := &MockSecretManager{
		GetSecretDataFunc: func(secretName string) (models.SecretData, error) {
			return models.SecretData{
				MongoDBURI: "mongodb://mock-uri",
			}, nil
		},
	}

	// Mock Mongo connect function
	originalMongoConnectFunc := DefaultMongoConnectFunc
	defer func() { DefaultMongoConnectFunc = originalMongoConnectFunc }() // Restore original after test
	DefaultMongoConnectFunc = func(ctx context.Context, uri string) (MongoClientInterface, error) {
		return &MockMongoClient{
			PingFunc: func(ctx context.Context, rp *readpref.ReadPref) error {
				return errors.New("failed to ping MongoDB")
			},
		}, nil
	}

	// Override SecretManagerFunc to return the mock SecretManager
	originalSecretManagerFunc := SecretManagerFunc
	defer func() { SecretManagerFunc = originalSecretManagerFunc }()
	SecretManagerFunc = func() secretsmanager.SecretManager {
		return mockSM
	}

	err := InitializeMongoDB()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ping MongoDB")
}

// func TestMongoClientWrapper_Ping(t *testing.T) {
// 	// Create a mock Mongo client and wrap it in MongoClientWrapper
// 	mockMongoClient := &mongo.Client{}
// 	wrapper := &MongoClientWrapper{Client: mockMongoClient}

// 	// Mock the Ping function of the actual client to simulate successful ping
// 	ctx := context.TODO()
// 	mockMongoClient.Ping(ctx, readpref.Primary())

// 	// Call Ping on the wrapper and check for success
// 	err := wrapper.Ping(ctx, readpref.Primary())
// 	assert.NoError(t, err)
// }

func TestMongoClientWrapper_Database(t *testing.T) {
	// Create a mock Mongo client and wrap it in MongoClientWrapper
	mockMongoClient := &mongo.Client{}
	wrapper := &MongoClientWrapper{Client: mockMongoClient}

	// Mock the Database function to return a basic mock database
	dbName := "testdb"
	// mockDatabase := &mongo.Database{}
	mockMongoClient.Database(dbName)

	// Call Database on the wrapper and check the returned value
	db := wrapper.Database(dbName)
	assert.NotNil(t, db)
	assert.Equal(t, dbName, db.Name())
}

func TestDefaultMongoConnectFunc(t *testing.T) {
	// Mock context and URI
	ctx := context.TODO()
	mockURI := "mongodb://mock-uri"

	// Call DefaultMongoConnectFunc
	clientInterface, err := DefaultMongoConnectFunc(ctx, mockURI)
	assert.NoError(t, err)
	assert.NotNil(t, clientInterface)

	// Check if clientInterface is of type MongoClientWrapper
	wrapper, ok := clientInterface.(*MongoClientWrapper)
	assert.True(t, ok)
	assert.NotNil(t, wrapper.Client)
}

// func TestSecretManagerFunc(t *testing.T) {
// 	// Call SecretManagerFunc and check the returned SecretManager
// 	secretManager := SecretManagerFunc()
// 	assert.NotNil(t, secretManager)

// 	// Type assertion to check if it returns the expected type
// 	_, ok := secretManager.(*secretsmanager.SecretManager)
// 	assert.True(t, ok)
// }


// func TestGetCollection(t *testing.T) {
// 	// Mock MongoClient to return a MockMongoDatabase wrapped in MongoDatabaseWrapper
// 	mockMongoClient := &MockMongoClient{
// 		DatabaseFunc: func(name string, opts ...*options.DatabaseOptions) *mongo.Database {
// 			return &mongo.Database{
// 				CollectionFunc: func(name string, opts ...*options.CollectionOptions) *mongo.Collection {
// 					return &mongo.Collection{} // Mock collection object
// 				},
// 			}
// 		},
// 	}
// 	MongoClient = mockMongoClient

// 	collectionName := "test-collection"
// 	collection := GetCollection(collectionName)

// 	assert.NotNil(t, collection)
// 	assert.Equal(t, collectionName, collection.Name())
// }
