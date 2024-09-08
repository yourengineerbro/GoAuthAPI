package auth

import (
	"errors"
	"net/http"
	"time"

	"GoAuthAPI/internal/model"
	"GoAuthAPI/internal/storage"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)
// type ServiceUtil interface {
// 	getClaims(tokenStr string) (*model.Claims, error)
// 	validateToken(tokenStr string) (*model.Claims, error)
// 	createToken(email string) (string, error)
// }

type Service struct {
	store  storage.UserStorage
	tokenBlacklist storage.TokenBlacklist
	jwtKey []byte
}

func NewService(store storage.UserStorage, tokenBlacklist storage.TokenBlacklist, jwtKey []byte) *Service {
	return &Service{store: store, tokenBlacklist: tokenBlacklist, jwtKey: jwtKey}
}

func (s *Service) DoesUserExists(email string) bool {
	exists := s.store.DoesUserExists(email)
	return exists
}

func (s *Service) CreateUser(user model.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return s.store.CreateUser(user)
}

func (s *Service) Authenticate(creds model.Credentials) (string, int, error) {
	user, err := s.store.GetUserByEmail(creds.Email)
	if err != nil {
		return "", http.StatusUnauthorized, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		return "", http.StatusUnauthorized, errors.New("invalid password")
	}

	token, err := s.CreateToken(creds.Email)
	if err != nil {
		return "", http.StatusInternalServerError, errors.New("Error creating token")
	}

	return token, http.StatusOK, nil
}

func (s *Service) RefreshToken(tokenStr string) (string, int, error) {

	claims, err := s.ValidateToken(tokenStr)
	if err != nil {
		return "", http.StatusUnauthorized, err
	}
	 
	s.RevokeToken(tokenStr)

	newToken, err := s.CreateToken(claims.Email)
	if err != nil {
		return "", http.StatusInternalServerError, errors.New("Error creating new token")
	}
	
	return newToken, http.StatusOK, nil
}

func (s *Service) RevokeToken(tokenStr string) {
	s.tokenBlacklist.BlacklistToken(tokenStr)
}

func (s *Service) GetClaims(tokenStr string) (*model.Claims, error) {
	claims := &model.Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}


func (s *Service) ValidateToken(tokenStr string) (*model.Claims, error) {
	claims := &model.Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})

	if err != nil {
		return nil, errors.New("No token found")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if s.tokenBlacklist.IsBlacklisted(tokenStr) {
		return nil, errors.New("token has been revoked")
	}
	
	return claims, nil
}

func (s *Service) CreateToken(email string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &model.Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtKey)
}

