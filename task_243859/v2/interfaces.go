
package main

type UserService interface {
    Create(user *User) error
    GetByID(id int) (*User, error)
    GetByUsername(username string) (*User, error)
    List() ([]User, error)
    Update(user *User) error
    Delete(id int) error
    Authenticate(username, password string) (*User, error)
}
