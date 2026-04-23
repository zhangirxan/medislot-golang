package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"medislot/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	user.ID = uuid.NewString()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(query,
		user.ID, user.Name, user.Email,
		user.PasswordHash, user.Role,
		user.CreatedAt, user.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return models.ErrEmailAlreadyTaken
	}
	return err
}

func (r *userRepository) GetByID(id string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, role, created_at, updated_at
	          FROM users WHERE id = $1`
	return r.scanUser(r.db.QueryRow(query, id))
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, role, created_at, updated_at
	          FROM users WHERE email = $1`
	return r.scanUser(r.db.QueryRow(query, email))
}

func (r *userRepository) GetAll() ([]*models.User, error) {
	query := `SELECT id, name, email, password_hash, role, created_at, updated_at
	          FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepository) Update(user *models.User) error {
	user.UpdatedAt = time.Now()
	query := `UPDATE users SET name=$1, email=$2, role=$3, updated_at=$4 WHERE id=$5`
	res, err := r.db.Exec(query, user.Name, user.Email, user.Role, user.UpdatedAt, user.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return models.ErrEmailAlreadyTaken
		}
		return err
	}
	return checkRowsAffected(res, models.ErrUserNotFound)
}

func (r *userRepository) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, models.ErrUserNotFound)
}


func (r *userRepository) scanUser(row *sql.Row) (*models.User, error) {
	u := &models.User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	return u, err
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), "unique") || contains(err.Error(), "duplicate")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func checkRowsAffected(res sql.Result, notFoundErr error) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return notFoundErr
	}
	return nil
}
