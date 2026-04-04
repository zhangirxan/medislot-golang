package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"medislot/internal/models"
	"medislot/internal/service"
	"medislot/pkg/middleware"
	"medislot/pkg/utils"
)

type UserHandler struct {
	userSvc service.UserService
}

func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userSvc.GetAll()
	if err != nil {
		slog.Error("GetAll users failed", "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, users)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	callerID := middleware.GetCallerID(c)
	callerRole := middleware.GetCallerRole(c)
	if callerRole != models.RoleAdmin && callerID != id {
		utils.ErrorResponse(c, models.ErrForbidden)
		return
	}

	user, err := h.userSvc.GetByID(id)
	if err != nil {
		slog.Warn("GetByID user not found", "id", id, "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	callerID := middleware.GetCallerID(c)
	callerRole := middleware.GetCallerRole(c)
	if callerRole != models.RoleAdmin && callerID != id {
		utils.ErrorResponse(c, models.ErrForbidden)
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if req.Role != "" && callerRole != models.RoleAdmin {
		utils.ErrorResponse(c, models.ErrForbidden)
		return
	}

	user, err := h.userSvc.Update(id, &req)
	if err != nil {
		slog.Warn("Update user failed", "id", id, "error", err)
		utils.ErrorResponse(c, err)
		return
	}

	slog.Info("user updated", "id", id, "caller", callerID)
	utils.SuccessResponse(c, http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.userSvc.Delete(id); err != nil {
		slog.Warn("Delete user failed", "id", id, "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	slog.Info("user deleted", "id", id)
	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "user deleted"})
}
