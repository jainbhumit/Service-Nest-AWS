package util

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"sync"
	"time"
)

const maxJWTSecretLen = 512

var (
	jwtSecret   []byte
	jwtSecretMu sync.RWMutex
)

// InitJWT sets the signing key from configuration. Must be called after config.Load().
func InitJWT(secret string) error {
	if secret == "" {
		return errors.New("jwt secret must not be empty")
	}
	if len(secret) > maxJWTSecretLen {
		return errors.New("jwt secret exceeds maximum length")
	}

	jwtSecretMu.Lock()
	jwtSecret = []byte(secret)
	jwtSecretMu.Unlock()
	return nil
}

func getJWTSecret() ([]byte, error) {
	jwtSecretMu.RLock()
	defer jwtSecretMu.RUnlock()
	if len(jwtSecret) == 0 {
		return nil, errors.New("jwt secret is not initialized")
	}
	return jwtSecret, nil
}

func GenerateJWT(userID string, role string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userID
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// VerifyJWT verifies the given JWT token.
func VerifyJWT(tokenString string) (*jwt.Token, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}
