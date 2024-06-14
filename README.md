
# Notification Service Project

## Overview

The Notification Service project is designed to facilitate real-time messaging and notifications between users. This project uses a combination of AWS services (Secrets Manager, Cognito), Kafka, MongoDB, and HTTP servers to manage and process notifications efficiently.

## Features

1. **User Authentication:** Uses AWS Cognito to validate JWT tokens for user authentication.
2. **Profile Validation:** Fetches and validates user profiles through a mock API.
3. **Kafka Integration:** Consumes messages from a Kafka topic and processes them as notifications.
4. **Notification Processing:** Processes incoming notifications and stores them in MongoDB.
5. **Real-time Notification Server:** A mock server to handle and display notifications dynamically.
6. **Secrets Management:** Retrieves sensitive information such as credentials and configuration details from AWS Secrets Manager.

## Project Structure

- **`auth/`**: Contains code for user authentication and profile validation.
- **`cmd/mockserver/`**: Implements a mock server to simulate notification handling.
- **`database/`**: Manages MongoDB connection and database interactions.
- **`kafka/`**: Consumes messages from Kafka and processes them.
- **`models/`**: Holds the struct definitions used across the project.
- **`secrets-manager/`**: Fetches secrets from AWS Secrets Manager.
- **`services/`**: Handles the logic for sending notifications to the UI.

## Prerequisites

1. **Go**: Make sure you have Go installed on your machine.
2. **AWS CLI**: Install and configure the AWS CLI to access AWS services.
3. **MongoDB**: Have access to a MongoDB instance.
4. **Kafka**: Ensure a Kafka broker is available for consuming messages.
5. **AWS Secrets Manager**: Store the required secrets under the secret name `notifsecrets`.

### Required Secrets in `notifsecrets`

```json
{
  "mongodbcreds": {
    "username": "your_mongodb_username",
    "password": "your_mongodb_password",
    "mongodbURI": "mongodb+srv://<username>:<password>@messages-cluster.utdovdn.mongodb.net/PROJECT0?retryWrites=true&w=majority"
  },
  "COGNITO_TOKEN": "your_cognito_token",
  "region": "us-east-1",
  "userPoolId": "us-east-1_icXeg2eiv",
  "profileURL": "https://fc76bbfd-f022-43e1-bb9d-355a53bae2cc.mock.pstmn.io",
  "kafkaBroker": "34.224.79.8:9092",
  "kafkaTopic": "chat-topic-46",
  "notificationEndpoint": "http://localhost:8081/notify"
}
```

Replace the placeholder values with actual credentials and configurations specific to your setup.

## Running the Project

### Step 1: Run the Mock Notification Server

The mock server simulates a notification service that processes and displays notifications.

1. Navigate to the `cmd/mockserver` directory.
2. Run the mock server using the following command:
   ```bash
   go run cmd/mockserver/mockserver.go
   ```
3. The server will start on `http://localhost:8081`. This server listens for incoming notification requests and displays them dynamically.

### Step 2: Start the Main Application

The main application handles user authentication, profile validation, and consumes Kafka messages.

1. Ensure you have the secrets stored in AWS Secrets Manager under the name `notifsecrets`.
2. Run the main application from the root directory with the following command:
   ```bash
   go run main.go
   ```
3. The application will:
   - Validate the user's Cognito token.
   - Fetch and validate user profiles using the profile URL.
   - Initialize MongoDB connection.
   - Consume messages from the specified Kafka topic.
   - Process these messages and send notifications.

## Dependencies

- **AWS SDK for Go**: For interacting with AWS services such as Secrets Manager and Cognito.
- **Sarama**: A Go library for interacting with Kafka.
- **MongoDB Go Driver**: To connect and interact with MongoDB.
- **Standard Go Libraries**: For HTTP handling, JSON encoding/decoding, and other core functionalities.

## Configuration

Ensure your AWS CLI is configured to access AWS Secrets Manager. The region and credentials should be set up using:

```bash
aws configure
```