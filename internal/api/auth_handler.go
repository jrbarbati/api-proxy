package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthTokenRequest struct {
	GrantType    string `json:"grant_type"`
	KindeToken   string `json:"token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (server *Server) handleOAuth(w http.ResponseWriter, r *http.Request) {
	authRequest, err := decodeJSON[AuthTokenRequest](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	switch authRequest.GrantType {
	case "client_credentials":
		handleClientCredentials(server, w, r, authRequest)
	case "kinde_token":
		handleKindeToken(w, r, authRequest)
	default:
		http.Error(w, "invalid grant type", http.StatusBadRequest)
	}
}

func handleClientCredentials(server *Server, w http.ResponseWriter, r *http.Request, authRequest *AuthTokenRequest) {
	account, err := findServiceAccount(server.serviceAccountRepository, authRequest.ClientID, authRequest.ClientSecret)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	if account == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, err := issueTokenForServiceAccount(server.jwtSigningSecret, account)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, accessToken, http.StatusOK)
}

func handleKindeToken(w http.ResponseWriter, r *http.Request, authRequest *AuthTokenRequest) {
	http.Error(w, "kinde token not supported", http.StatusNotImplemented)
}

func issueTokenForServiceAccount(signingSecret string, serviceAccount *model.ServiceAccount) (*AccessToken, error) {
	expiresIn := 3600

	claims := jwt.MapClaims{
		"sub":    serviceAccount.ID,
		"org_id": serviceAccount.OrgID,
		"type":   "external",
		"iss":    "api-proxy",
		"exp":    time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(signingSecret))

	if err != nil {
		return nil, err
	}

	return &AccessToken{
		AccessToken: signed,
		ExpiresIn:   expiresIn,
	}, nil
}

func findServiceAccount(serviceAccountRepository *repository.ServiceAccountRepository, clientId, clientSecret string) (*model.ServiceAccount, error) {
	serviceAccount, err := serviceAccountRepository.FindByClientID(clientId)

	if err != nil {
		return nil, err
	}

	savedSecret := "$2a$10$Zs3OyuUJSShI5qQiM/SDQuqdXBEhpfqG4h9A4gUC/StpJlVz9TUUa" // Preventing side-channel attack
	if serviceAccount != nil {
		savedSecret = serviceAccount.ClientSecret
	}

	if !matches(savedSecret, clientSecret) {
		return nil, nil
	}

	return serviceAccount, nil
}

func matches(savedSecret, requestSecret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(savedSecret), []byte(requestSecret)) == nil
}

func validateKindeToken() (bool, error) {
	return false, nil
}
