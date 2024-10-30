package main

import (
	"database/sql"
	_ "database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type SQLUserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) UserService {
	return &SQLUserService{db: db}
}

func (s *SQLUserService) Create(user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result, err := s.db.Exec(
		"INSERT INTO users (username, password, email) VALUES (?, ?, ?)",
		user.Username,
		hashedPassword,
		user.Email,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

func (s *SQLUserService) GetByID(id int) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(
		"SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *SQLUserService) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(
		"SELECT id, username, password, email, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *SQLUserService) List() ([]User, error) {
	rows, err := s.db.Query("SELECT id, username, email, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *SQLUserService) Delete(id int) error {
	result, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (s *SQLUserService) Authenticate(username, password string) (*User, error) {
	user, err := s.GetByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
