
package main

import "errors"

var (
    ErrUserNotFound      = errors.New("user not found")
    ErrDuplicateUsername = errors.New("username already exists")
    ErrDuplicateEmail    = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
)
