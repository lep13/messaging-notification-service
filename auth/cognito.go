package auth

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/dgrijalva/jwt-go"
)

var (
	cognitoRegion = "your-region"       // AWS region where Cognito is set up
	cognitoPoolID = "your-user-pool-id" // Cognito User Pool ID
	jwksURL       = fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", cognitoRegion, cognitoPoolID)
	cachedJWKS    *JWKS
	jwksSync      sync.Mutex
)

// JWKS represents the JSON Web Key Set
type JWKS struct {
	Keys []json.RawMessage `json:"keys"`
}

// ValidateUser validates the user by decoding and verifying the JWT
func ValidateUser(token string) (bool, error) {
	// Get JWKS for token validation
	jwks, err := getJWKS()
	if err != nil {
		return false, err
	}

	// Parse the JWT token
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method is RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}

		kid := token.Header["kid"].(string)

		// Find the appropriate key from the JWKS
		for _, key := range jwks.Keys {
			rsaPublicKey, err := parseRSAPublicKey(key)
			if err != nil {
				continue
			}
			if jwtToken.Header["kid"] == rsaPublicKey.KeyID {
				return rsaPublicKey.PublicKey, nil
			}
		}
		return nil, errors.New("unable to find key")
	})

	if err != nil || !jwtToken.Valid {
		return false, err
	}

	// Additional validation can be added here
	return true, nil
}

// getJWKS retrieves and caches the JWKS from the specified URL
func getJWKS() (*JWKS, error) {
	jwksSync.Lock()
	defer jwksSync.Unlock()

	if cachedJWKS != nil {
		return cachedJWKS, nil
	}

	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	cachedJWKS = &jwks
	return &jwks, nil
}

// parseRSAPublicKey parses an RSA public key from the JWKS
func parseRSAPublicKey(key json.RawMessage) (*rsa.PublicKey, error) {
	var rsaKey struct {
		KeyID     string `json:"kid"`
		PublicKey *rsa.PublicKey
	}
	err := json.Unmarshal(key, &rsaKey)
	if err != nil {
		return nil, err
	}
	return rsaKey.PublicKey, nil
}
