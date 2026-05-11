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

type userRepository struct{ db *sql.DB }

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	user.ID = uuid.NewString()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT INTO users (id, name, email, password_hash, role, doctor_type, medical_notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		user.ID, user.Name, user.Email,
		user.PasswordHash, user.Role, user.DoctorType,
		user.MedicalNotes,
		user.CreatedAt, user.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return models.ErrEmailAlreadyTaken
	}
	return err
}

func (r *userRepository) GetByID(id string) (*models.User, error) {
	return r.scanUser(r.db.QueryRow(`
		SELECT id, name, email, password_hash, role, doctor_type, medical_notes, created_at, updated_at
		FROM users WHERE id = $1`, id))
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	return r.scanUser(r.db.QueryRow(`
		SELECT id, name, email, password_hash, role, doctor_type, medical_notes, created_at, updated_at
		FROM users WHERE email = $1`, email))
}

func (r *userRepository) GetAll() ([]*models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, name, email, password_hash, role, doctor_type, medical_notes, created_at, updated_at
		FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.PasswordHash,
			&u.Role, &u.DoctorType, &u.MedicalNotes,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepository) Update(user *models.User) error {
	user.UpdatedAt = time.Now()
	res, err := r.db.Exec(`
		UPDATE users SET name=$1, email=$2, role=$3, doctor_type=$4, medical_notes=$5, updated_at=$6
		WHERE id=$7`,
		user.Name, user.Email, user.Role,
		user.DoctorType, user.MedicalNotes,
		user.UpdatedAt, user.ID,
	)
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
	err := row.Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash,
		&u.Role, &u.DoctorType, &u.MedicalNotes,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	return u, err
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	for i := 0; i <= len(s)-6; i++ {
		if s[i:i+6] == "unique" {
			return true
		}
	}
	for i := 0; i <= len(s)-9; i++ {
		if s[i:i+9] == "duplicate" {
			return true
		}
	}
	return false
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
