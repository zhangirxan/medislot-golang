package service

import (
	"medislot/internal/models"
	"medislot/internal/repository"
	"medislot/pkg/utils"
)

type UserService interface {
	Register(req *models.RegisterRequest) (*models.User, error)
	Login(req *models.LoginRequest, jwtSecret string, expiryHours int) (*models.LoginResponse, error)
	GetByID(id string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Update(id string, req *models.UpdateUserRequest) (*models.User, error)
	Delete(id string) error
}

type userService struct{ repo repository.UserRepository }

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(req *models.RegisterRequest) (*models.User, error) {
	if err := validateDoctorTypeForRole(req.Role, req.DoctorType); err != nil {
		return nil, err
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
		DoctorType:   req.DoctorType,
		MedicalNotes: req.MedicalNotes,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Login(req *models.LoginRequest, jwtSecret string, expiryHours int) (*models.LoginResponse, error) {
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return nil, models.ErrInvalidCredentials
	}
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, models.ErrInvalidCredentials
	}
	token, err := utils.GenerateToken(user, jwtSecret, expiryHours)
	if err != nil {
		return nil, err
	}
	return &models.LoginResponse{Token: token, User: user}, nil
}

func (s *userService) GetByID(id string) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetAll() ([]*models.User, error) {
	return s.repo.GetAll()
}

func (s *userService) Update(id string, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
		if req.Role != models.RoleDoctor && req.DoctorType == nil {
			user.DoctorType = nil
		}
	}
	if req.DoctorType != nil {
		user.DoctorType = req.DoctorType
	}
	if req.MedicalNotes != nil {
		user.MedicalNotes = *req.MedicalNotes
	}
	if err := validateDoctorTypeForRole(user.Role, user.DoctorType); err != nil {
		return nil, err
	}
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Delete(id string) error {
	return s.repo.Delete(id)
}

func validateDoctorTypeForRole(role models.Role, doctorType *models.DoctorType) error {
	if doctorType != nil && !models.IsValidDoctorType(*doctorType) {
		return models.ErrInvalidDoctorType
	}
	if role == models.RoleDoctor {
		if doctorType == nil {
			return models.ErrDoctorTypeRequired
		}
		return nil
	}
	if doctorType != nil {
		return models.ErrDoctorTypeOnlyForDoctor
	}
	return nil
}
