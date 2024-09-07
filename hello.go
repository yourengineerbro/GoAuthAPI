package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
)

var users = map[string]string{}
var jwtKey = []byte("your_secret_key")
var blacklistedTokens = make(map[string]bool)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

type Response struct {
	Message string `json:"message"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/api/auth/signup", SignUpHandler)
	r.Post("/api/auth/signin", SignInHandler)
	r.Post("/api/auth/refresh", RefreshHandler)
	r.Post("/api/auth/revoke", RevokeHandler)
	r.Get("/api/protected", AuthMiddleware(ProtectedHandler))

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !isValidEmail(creds.Email) {
		sendErrorResponse(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if _, exists := users[creds.Email]; exists {
		sendErrorResponse(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		sendErrorResponse(w, "Error while hashing password", http.StatusInternalServerError)
		return
	}

	users[creds.Email] = string(hashedPassword)
	sendJSONResponse(w, Response{Message: "User registered successfully"}, http.StatusCreated)
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	storedPassword, ok := users[creds.Email]
	if !ok {
		sendErrorResponse(w, "User not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password)); err != nil {
		sendErrorResponse(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := createToken(creds.Email)
	if err != nil {
		sendErrorResponse(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, token)
	sendJSONResponse(w, Response{Message: "Logged in successfully"}, http.StatusOK)
}

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
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

	if blacklistedTokens[tokenStr] {
		sendErrorResponse(w, "Invalid Token", http.StatusUnauthorized)
		return
	}

	blacklistedTokens[tokenStr] = true

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			sendErrorResponse(w, "Invalid token signature", http.StatusUnauthorized)
			return
		}
		sendErrorResponse(w, "Error parsing token", http.StatusBadRequest)
		return
	}

	if !tkn.Valid {
		sendErrorResponse(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// if time.Until(time.Unix(claims.ExpiresAt, 0)) > 30*time.Second {
	// 	sendErrorResponse(w, "Token is not expired yet", http.StatusBadRequest)
	// 	return
	// }

	newToken, err := createToken(claims.Email)
	if err != nil {
		sendErrorResponse(w, "Error creating new token", http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, newToken)
	sendJSONResponse(w, Response{Message: "Token refreshed successfully"}, http.StatusOK)
}

func RevokeHandler(w http.ResponseWriter, r *http.Request) {
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
	blacklistedTokens[tokenStr] = true

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Path:     "/api",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	sendJSONResponse(w, Response{Message: "Logged out successfully"}, http.StatusOK)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if blacklistedTokens[tokenStr] {
			sendErrorResponse(w, "Token has been revoked", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				sendErrorResponse(w, "Invalid token signature", http.StatusUnauthorized)
				return
			}
			sendErrorResponse(w, "Error parsing token", http.StatusBadRequest)
			return
		}

		if !tkn.Valid {
			sendErrorResponse(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie("token")
	claims := &Claims{}
	jwt.ParseWithClaims(c.Value, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	sendJSONResponse(w, Response{Message: fmt.Sprintf("Welcome %s! This is a protected resource.", claims.Email)}, http.StatusOK)
}

func createToken(email string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
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

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
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