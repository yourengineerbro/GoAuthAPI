package storage

import (
	"errors"
	"GoAuthAPI/internal/model"
	"sync"
)

type UserStorage interface {
	CreateUser(user model.User) error
	DoesUserExists(email string) bool
	GetUserByEmail(email string) (model.User, error)
}

type MemoryStorage struct {
	users map[string]model.User
	mutex sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users: make(map[string]model.User),
	}
}

func (m *MemoryStorage) CreateUser(user model.User) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.users[user.Email] = user
	return nil
}

func (m *MemoryStorage) DoesUserExists(email string) bool {
	if _, exists := m.users[email]; exists {
		return true
	}
	return false
}

func (m *MemoryStorage) GetUserByEmail(email string) (model.User, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	user, exists := m.users[email]
	if !exists {
		return model.User{}, errors.New("user not found")
	}

	return user, nil
}


// todo : blacklist token in different file
// Blacklist Token

type TokenBlacklist interface {
	BlacklistToken(token string)
	IsBlacklisted(token string) bool
}

type MemoryTokenBlacklist struct {
	blacklist map[string]struct{}
	mutex     sync.RWMutex
}

func NewMemoryTokenBlacklist() *MemoryTokenBlacklist {
	blacklist := &MemoryTokenBlacklist{
		blacklist: make(map[string]struct{}),
	}
	return blacklist
}

func (m *MemoryTokenBlacklist) BlacklistToken(token string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.blacklist[token] = struct{}{}
}

func (m *MemoryTokenBlacklist) IsBlacklisted(token string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.blacklist[token]
	return exists
}

