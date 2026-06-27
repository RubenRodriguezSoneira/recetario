package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"recipe-app/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user *models.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	now := time.Now()

	query := `
		INSERT INTO users (id, email, username, first_name, last_name, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Password,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id string) (*models.User, error) {
	query := `
		SELECT id, email, username, first_name, last_name, password_hash, avatar_url, created_at, updated_at
		FROM users
		WHERE id = ?`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, username, first_name, last_name, password_hash, avatar_url, created_at, updated_at
		FROM users
		WHERE email = ?`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, email, username, first_name, last_name, password_hash, avatar_url, created_at, updated_at
		FROM users
		WHERE username = ?`

	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users 
		SET email = ?, username = ?, first_name = ?, last_name = ?, avatar_url = ?, updated_at = ?
		WHERE id = ?`

	_, err := r.db.Exec(
		query,
		user.Email,
		user.Username,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(userID, password string) error {
	query := `
		UPDATE users 
		SET password_hash = ?, updated_at = ?
		WHERE id = ?`

	_, err := r.db.Exec(query, password, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeleteUser deletes a user and all their data
func (r *UserRepository) DeleteUser(id string) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)"

	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// UsernameExists checks if a username already exists
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)"

	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}
