package main

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserService interface {
	Create(user *User) error
	GetByID(id int) (*User, error)
	GetByUsername(username string) (*User, error)
	List() ([]User, error)
	Delete(id int) error
	Authenticate(username, password string) (*User, error)
}
