
package main

import (
    "time"
)

type User struct {
    ID        int       `json:"id"`
    Username  string    `json:"username" binding:"required"`
    Password  string    `json:"password,omitempty" binding:"required,min=6"`
    Email     string    `json:"email" binding:"required,email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
