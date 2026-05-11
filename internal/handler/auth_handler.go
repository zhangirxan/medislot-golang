package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"medislot/internal/models"
	"medislot/internal/service"
	"medislot/pkg/utils"
)

type AuthHandler struct {
	userSvc     service.UserService
	jwtSecret   string
	expiryHours int
}

func NewAuthHandler(userSvc service.UserService, jwtSecret string, expiryHours int) *AuthHandler {
	return &AuthHandler{userSvc: userSvc, jwtSecret: jwtSecret, expiryHours: expiryHours}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}
	user, err := h.userSvc.Register(&req)
	if err != nil {
		utils.ErrorResponse(c, err)
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}
	resp, err := h.userSvc.Login(&req, h.jwtSecret, h.expiryHours)
	if err != nil {
		utils.ErrorResponse(c, err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, resp)
}
