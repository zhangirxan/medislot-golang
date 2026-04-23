package models

import "time"

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleDoctor  Role = "doctor"
	RolePatient Role = "patient"
)

type User struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         Role      `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}


type RegisterRequest struct {
	Name     string `json:"name"     binding:"required,min=2,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     Role   `json:"role"     binding:"required,oneof=admin doctor patient"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"  binding:"omitempty,min=2,max=100"`
	Email string `json:"email" binding:"omitempty,email"`
	Role  Role   `json:"role"  binding:"omitempty,oneof=admin doctor patient"`
}
