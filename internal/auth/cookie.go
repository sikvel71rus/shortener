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
)

const (
	CookieName = "user_token"
	secretKey  = "supersecretkey"
)

var (
	ErrInvalidToken = errors.New("invalid auth token")
	ErrEmptyUserID  = errors.New("empty user id")
)

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

	payload := base64.RawURLEncoding.EncodeToString([]byte(userID))
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payload + "." + signature, nil
}

func ParseUserID(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", ErrInvalidToken
	}

	payload := parts[0]
	mac := hmac.New(sha256.New, []byte(secretKey))
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

func generateUserID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
