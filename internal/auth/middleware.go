package auth

import (
	"context"
	"net/http"
	"encoding/json"

	"GoAuthAPI/internal/model"
)

// Key type for context values
type contextKey string

// Define a key for claims in the context
const claimsContextKey contextKey = "claims"

func Middleware(service *Service) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("token")
			if err != nil {
				if err == http.ErrNoCookie {
					sendMiddleWareErrorResponse(w, "No token found", http.StatusUnauthorized)
					return
				}
				sendMiddleWareErrorResponse(w, "Error reading cookie", http.StatusBadRequest)
				return
			}

			tokenStr := c.Value
			claims, err := service.ValidateToken(tokenStr)
			if err != nil {
				sendMiddleWareErrorResponse(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := setClaimsToContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// setClaimsToContext adds the provided claims to the context
func setClaimsToContext(ctx context.Context, claims *model.Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// GetClaimsFromContext retrieves the claims from the context
// func GetClaimsFromContext(ctx context.Context) *model.Claims {
// 	if claims, ok := ctx.Value(claimsContextKey).(*model.Claims); ok {
// 		return claims
// 	}
// 	return nil
// }

// Helper function to send error responses
func sendMiddleWareErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}