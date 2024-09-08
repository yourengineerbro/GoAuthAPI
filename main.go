package main

import (
	"fmt"
	"log"
	"net/http"

	"GoAuthAPI/config"
	"GoAuthAPI/internal/auth"
	"GoAuthAPI/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()
	store := storage.NewMemoryStorage()
	tokenBlacklist := storage.NewMemoryTokenBlacklist()
	authService := auth.NewService(store, tokenBlacklist, cfg.JWTKey)
	authHandler := auth.NewHandler(authService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/api/auth/signup", authHandler.SignUp)
	r.Post("/api/auth/signin", authHandler.SignIn)
	r.Post("/api/auth/refresh", authHandler.Refresh)
	r.Post("/api/auth/revoke", authHandler.Revoke)
	r.Get("/api/protected", auth.Middleware(authService)(authHandler.Protected))

	fmt.Printf("Server is running on port %s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}