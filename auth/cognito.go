package auth

import (
    "context"
    "log"

    "github.com/ShreerajShettyK/cognitoJwtAuthenticator"
    secretsmanager "github.com/lep13/messaging-notification-service/secrets-manager"
)

type TokenValidatorFunc func(ctx context.Context, region, userPoolID, tokenString string) (*cognitoJwtAuthenticator.AWSCognitoClaims, error)

var DefaultTokenValidator TokenValidatorFunc = cognitoJwtAuthenticator.ValidateToken

// ValidateCognitoToken validates the provided JWT token with AWS Cognito using cognitoJwtAuthenticator
func ValidateCognitoToken(ctx context.Context, tokenString string, sm secretsmanager.SecretManager, validator TokenValidatorFunc) (*cognitoJwtAuthenticator.AWSCognitoClaims, error) {
    // Fetch the region and userPoolId from secrets manager
    secretName := "notifsecrets"
    secrets, err := sm.GetSecretData(secretName)
    if err != nil {
        log.Printf("Error retrieving secrets: %v", err)
        return nil, err
    }

    region := secrets.Region
    userPoolId := secrets.UserPoolID

    claims, err := validator(ctx, region, userPoolId, tokenString)
    if err != nil {
        log.Printf("Token validation error: %s", err)
        return nil, err
    }
    log.Println("Token is valid")
    log.Printf("Claims: %+v", claims)
    return claims, nil
}
