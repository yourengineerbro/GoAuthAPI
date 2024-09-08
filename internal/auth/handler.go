package auth

import (
	"fmt"
	"encoding/json"
	"net/http"
	"time"

	"GoAuthAPI/internal/model"
	"GoAuthAPI/pkg/validator"
)

type Handler struct {
	service *Service
}

type Response struct {
    Message string `json:"message"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !validator.IsValidEmail(user.Email) {
		sendErrorResponse(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if exists := h.service.DoesUserExists(user.Email); exists {
		sendErrorResponse(w, "User already exists", http.StatusConflict)
		return
	}

	if err := h.service.CreateUser(user); err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, Response{Message: "User registered successfully"}, http.StatusCreated)
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	var creds model.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if !validator.IsValidEmail(creds.Email) {
		sendErrorResponse(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	token, status, err := h.service.Authenticate(creds)
	if err != nil {
		sendErrorResponse(w, err.Error(), status)
		return
	}

	setTokenCookie(w, token)
	sendJSONResponse(w, Response{Message: "Logged in successfully"}, status)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			sendErrorResponse(w, "No token found", http.StatusUnauthorized)
			return
		}
		sendErrorResponse(w, "Error reading cookie", http.StatusBadRequest)
		return
	}
	tokenStr := c.Value
	newToken, status, err := h.service.RefreshToken(tokenStr)
	if err != nil {
		sendErrorResponse(w, err.Error(), status)
		return
	}
	setTokenCookie(w, newToken)
	sendJSONResponse(w, Response{Message: "Token refreshed successfully"}, http.StatusOK)
}



func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			sendErrorResponse(w, "No token found", http.StatusUnauthorized)
			return
		}
		sendErrorResponse(w, "Error reading cookie", http.StatusBadRequest)
		return
	}

	tokenStr := c.Value
	_, err = h.service.ValidateToken(tokenStr)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.service.RevokeToken(tokenStr)
	setTokenCookie(w, "");
	sendJSONResponse(w, Response{Message: "Logged out successfully"}, http.StatusOK)
}

func (h *Handler) Protected(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie("token")
	claims, err := h.service.GetClaims(c.Value)
    if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusUnauthorized)
		return
	}
	sendJSONResponse(w, Response{Message: fmt.Sprintf("Welcome %s! This is a protected resource.", claims.Email)}, http.StatusOK)
}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/api",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
	})
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{Message: message})
}

func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
