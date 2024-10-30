// mock_service.go
package main

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type MockUserService struct {
	users  map[int]*User
	nextID int
}

func NewMockUserService() UserService {
	return &MockUserService{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (m *MockUserService) Create(user *User) error {
	// Check for duplicate username
	for _, existingUser := range m.users {
		if existingUser.Username == user.Username {
			return ErrDuplicateUsername
		}
		if existingUser.Email == user.Email {
			return ErrDuplicateEmail
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Set user fields
	user.ID = m.nextID
	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Store user
	m.users[user.ID] = user
	m.nextID++

	return nil
}

func (m *MockUserService) GetByID(id int) (*User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Return a copy to prevent modification of internal state
	userCopy := *user
	userCopy.Password = "" // Remove password from response
	return &userCopy, nil
}

func (m *MockUserService) GetByUsername(username string) (*User, error) {
	for _, user := range m.users {
		if user.Username == username {
			userCopy := *user
			return &userCopy, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MockUserService) List() ([]User, error) {
	users := make([]User, 0, len(m.users))
	for _, user := range m.users {
		userCopy := *user
		userCopy.Password = "" // Remove password from response
		users = append(users, userCopy)
	}
	return users, nil
}

func (m *MockUserService) Update(user *User) error {
	existingUser, exists := m.users[user.ID]
	if !exists {
		return ErrUserNotFound
	}

	// Check for duplicate username/email with other users
	for id, u := range m.users {
		if id != user.ID {
			if u.Username == user.Username {
				return ErrDuplicateUsername
			}
			if u.Email == user.Email {
				return ErrDuplicateEmail
			}
		}
	}

	// Update fields
	existingUser.Username = user.Username
	existingUser.Email = user.Email
	existingUser.UpdatedAt = time.Now()

	// Update password if provided
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		existingUser.Password = string(hashedPassword)
	}

	return nil
}

func (m *MockUserService) Delete(id int) error {
	if _, exists := m.users[id]; !exists {
		return ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserService) Authenticate(username, password string) (*User, error) {
	user, err := m.GetByUsername(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user.Password = "" // Remove password from response
	return user, nil
}

// Add these helper functions to main_test.go
func createTestUser(svc UserService) (*User, error) {
	user := &User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}
	err := svc.Create(user)
	return user, err
}

func clearMockData(svc UserService) {
	if mockSvc, ok := svc.(*MockUserService); ok {
		mockSvc.users = make(map[int]*User)
		mockSvc.nextID = 1
	}
}
