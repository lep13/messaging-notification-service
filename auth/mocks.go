package auth

import (
	"github.com/lep13/messaging-notification-service/models"
)

// MockSecretsManager is used in both files in auth dir
type MockSecretsManager struct {
	ProfileURL string
	Region     string
	UserPoolID string
}

func (m *MockSecretsManager) GetSecretData(secretName string) (models.SecretData, error) {
	return models.SecretData{
		ProfileURL: m.ProfileURL,
		Region:     m.Region,
		UserPoolID: m.UserPoolID,
	}, nil
}
