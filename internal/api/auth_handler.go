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

type InternalAuthTokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type AuthHandler struct {
	jwtSigningSecret         string
	adminJwtSigningSecret    string
	serviceAccountRepository *repository.ServiceAccountRepository
	internalUserRepository   *repository.InternalUserRepository
}

func NewAuthHandler(
	jwtSigningSecret string,
	adminJwtSigningSecret string,
	serviceAccountRepository *repository.ServiceAccountRepository,
	internalUserRepository *repository.InternalUserRepository,
) *AuthHandler {
	return &AuthHandler{
		jwtSigningSecret:         jwtSigningSecret,
		adminJwtSigningSecret:    adminJwtSigningSecret,
		serviceAccountRepository: serviceAccountRepository,
		internalUserRepository:   internalUserRepository,
	}
}

func (ah *AuthHandler) handleInternalOAuth(w http.ResponseWriter, r *http.Request) {
	authRequest, err := decodeJSON[InternalAuthTokenRequest](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	ah.handleInternalCredentials(w, r, authRequest)
}

func (ah *AuthHandler) handleOAuth(w http.ResponseWriter, r *http.Request) {
	authRequest, err := decodeJSON[AuthTokenRequest](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	switch authRequest.GrantType {
	case "client_credentials":
		ah.handleClientCredentials(w, r, authRequest)
	case "kinde_token":
		ah.handleKindeToken(w, r, authRequest)
	default:
		http.Error(w, "invalid grant type", http.StatusBadRequest)
	}
}

func (ah *AuthHandler) handleInternalCredentials(w http.ResponseWriter, r *http.Request, authRequest *InternalAuthTokenRequest) {
	user, err := ah.findInternalUser(authRequest.Email, authRequest.Password)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, err := ah.issueTokenForUser(user)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, accessToken, http.StatusOK)
}

func (ah *AuthHandler) handleClientCredentials(w http.ResponseWriter, r *http.Request, authRequest *AuthTokenRequest) {
	account, err := ah.findServiceAccount(authRequest.ClientID, authRequest.ClientSecret)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	if account == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, err := ah.issueTokenForServiceAccount(account)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, accessToken, http.StatusOK)
}

func (ah *AuthHandler) handleKindeToken(w http.ResponseWriter, r *http.Request, authRequest *AuthTokenRequest) {
	http.Error(w, "kinde token not supported", http.StatusNotImplemented)
}

func (ah *AuthHandler) issueTokenForUser(user *model.InternalUser) (*AccessToken, error) {
	var expiresIn = 3600

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"type": "internal",
		"iss":  "api-proxy",
		"exp":  time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(ah.adminJwtSigningSecret))

	if err != nil {
		return nil, err
	}

	return &AccessToken{
		AccessToken: signed,
		ExpiresIn:   expiresIn,
	}, nil
}

func (ah *AuthHandler) issueTokenForServiceAccount(serviceAccount *model.ServiceAccount) (*AccessToken, error) {
	var expiresIn = 3600

	claims := jwt.MapClaims{
		"sub":    serviceAccount.ID,
		"org_id": serviceAccount.OrgID,
		"type":   "external",
		"iss":    "api-proxy",
		"exp":    time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(ah.jwtSigningSecret))

	if err != nil {
		return nil, err
	}

	return &AccessToken{
		AccessToken: signed,
		ExpiresIn:   expiresIn,
	}, nil
}

func (ah *AuthHandler) findInternalUser(email, password string) (*model.InternalUser, error) {
	user, err := ah.internalUserRepository.FindByEmail(email)

	if err != nil {
		return nil, err
	}

	// TODO: A user with the correct password is being denied here

	savedSecret := "$2a$10$Zs3OyuUJSShI5qQiM/SDQuqdXBEhpfqG4h9A4gUC/StpJlVz9TUUa" // Preventing side-channel attack
	if user != nil {
		savedSecret = user.Password
	}

	if !matches(savedSecret, password) {
		return nil, nil
	}

	return user, nil
}

func (ah *AuthHandler) findServiceAccount(clientId, clientSecret string) (*model.ServiceAccount, error) {
	account, err := ah.serviceAccountRepository.FindByClientID(clientId)

	if err != nil {
		return nil, err
	}

	savedSecret := "$2a$10$Zs3OyuUJSShI5qQiM/SDQuqdXBEhpfqG4h9A4gUC/StpJlVz9TUUa" // Preventing side-channel attack
	if account != nil {
		savedSecret = account.ClientSecret
	}

	if !matches(savedSecret, clientSecret) {
		return nil, nil
	}

	return account, nil
}

func matches(savedSecret, requestSecret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(savedSecret), []byte(requestSecret)) == nil
}

func validateKindeToken() (bool, error) {
	return false, nil
}
