// sql_service.go
package main

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
)

type SQLUserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) UserService {
	return &SQLUserService{
		db: db,
	}
}

func (s *SQLUserService) Create(user *User) error {
	// Begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check for duplicate username
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", user.Username).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateUsername
	}

	// Check for duplicate email
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", user.Email).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateEmail
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insert user
	result, err := tx.Exec(`
        INSERT INTO users (username, password, email, created_at, updated_at)
        VALUES (?, ?, ?, NOW(), NOW())
    `, user.Username, hashedPassword, user.Email)
	if err != nil {
		return err
	}

	// Get inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)

	return tx.Commit()
}

func (s *SQLUserService) GetByID(id int) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(`
        SELECT id, username, email, created_at, updated_at
        FROM users
        WHERE id = ?
    `, id).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *SQLUserService) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(`
        SELECT id, username, password, email, created_at, updated_at
        FROM users
        WHERE username = ?
    `, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *SQLUserService) List() ([]User, error) {
	rows, err := s.db.Query(`
        SELECT id, username, email, created_at, updated_at
        FROM users
        ORDER BY id ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *SQLUserService) Update(user *User) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if user exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", user.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	// Check for duplicate username (excluding current user)
	err = tx.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM users 
            WHERE username = ? AND id != ?
        )
    `, user.Username, user.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateUsername
	}

	// Check for duplicate email (excluding current user)
	err = tx.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM users 
            WHERE email = ? AND id != ?
        )
    `, user.Email, user.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateEmail
	}

	if user.Password != "" {
		// Update with new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`
            UPDATE users 
            SET username = ?, password = ?, email = ?, updated_at = NOW()
            WHERE id = ?
        `, user.Username, hashedPassword, user.Email, user.ID)
	} else {
		// Update without changing password
		_, err = tx.Exec(`
            UPDATE users 
            SET username = ?, email = ?, updated_at = NOW()
            WHERE id = ?
        `, user.Username, user.Email, user.ID)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLUserService) Delete(id int) error {
	result, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *SQLUserService) Authenticate(username, password string) (*User, error) {
	user, err := s.GetByUsername(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Don't return the password
	user.Password = ""
	return user, nil
}
