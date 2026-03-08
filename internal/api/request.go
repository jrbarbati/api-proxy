package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func decodeJSON[T any](r *http.Request) (*T, error) {
	var body T

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return &body, nil
}

func queryParam[T any](param string, r *http.Request, conversion func(string) (*T, error)) (*T, error) {
	if param == "" {
		return nil, nil
	}

	var result *T
	var err error

	if val := r.URL.Query().Get(param); val != "" {
		result, err = conversion(val)
	}

	return result, err
}

func toIntParam(val string) (*int, error) {
	id, err := strconv.Atoi(val)

	if err != nil {
		return nil, err
	}

	return &id, nil
}

func hashSecret(secret string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}
