package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func decodeJSON[T any](r *http.Request) (*T, error) {
	var body T

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return &body, nil
}

func hashSecret(secret string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}
