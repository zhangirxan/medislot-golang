package service

import (
	"database/sql"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"medislot-backend/internal/auth"
	"medislot-backend/internal/domain"
	"medislot-backend/internal/dto"
	"medislot-backend/internal/repository"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(req dto.RegisterRequest) (string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	existingUser, err := s.userRepo.GetByEmail(email)
	if err == nil && existingUser != nil {
		return "", errors.New("user with this email already exists")
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	role := domain.RolePatient
	if req.Role != "" {
		switch req.Role {
		case string(domain.RoleAdmin):
			role = domain.RoleAdmin
		case string(domain.RoleDoctor):
			role = domain.RoleDoctor
		case string(domain.RolePatient):
			role = domain.RolePatient
		default:
			return "", errors.New("invalid role")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &domain.User{
		FullName:     req.FullName,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return "", err
	}

	token, err := auth.GenerateToken(user.ID, string(user.Role), s.jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (string, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("invalid email or password")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	token, err := auth.GenerateToken(user.ID, string(user.Role), s.jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) GetMe(userID int64) (*domain.User, error) {
	return s.userRepo.GetByID(userID)
}