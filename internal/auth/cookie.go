package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
)

const CookieName = "user_token"

var (
	ErrInvalidToken = errors.New("invalid auth token")
	ErrEmptyUserID  = errors.New("empty user id")
	ErrEmptySecret  = errors.New("empty auth secret")

	secretMu  sync.RWMutex
	secretKey string
)

func SetSecret(secret string) error {
	if strings.TrimSpace(secret) == "" {
		return ErrEmptySecret
	}

	secretMu.Lock()
	defer secretMu.Unlock()
	secretKey = secret

	return nil
}

func NewSignedCookie() (*http.Cookie, string, error) {
	userID, err := generateUserID()
	if err != nil {
		return nil, "", err
	}

	token, err := BuildToken(userID)
	if err != nil {
		return nil, "", err
	}

	return &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	}, userID, nil
}

func BuildToken(userID string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", ErrEmptyUserID
	}

	secret, err := getSecret()
	if err != nil {
		return "", err
	}

	payload := base64.RawURLEncoding.EncodeToString([]byte(userID))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payload + "." + signature, nil
}

func ParseUserID(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", ErrInvalidToken
	}

	secret, err := getSecret()
	if err != nil {
		return "", err
	}

	payload := parts[0]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := mac.Sum(nil)

	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ErrInvalidToken
	}

	if !hmac.Equal(signature, expected) {
		return "", ErrInvalidToken
	}

	userIDBytes, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", ErrInvalidToken
	}

	userID := string(userIDBytes)
	if strings.TrimSpace(userID) == "" {
		return "", ErrEmptyUserID
	}

	return userID, nil
}

func getSecret() (string, error) {
	secretMu.RLock()
	defer secretMu.RUnlock()

	if strings.TrimSpace(secretKey) == "" {
		return "", ErrEmptySecret
	}

	return secretKey, nil
}

func generateUserID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
