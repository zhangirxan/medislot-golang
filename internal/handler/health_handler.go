package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"medislot-backend/internal/service"
)

type HealthHandler struct {
	service *service.HealthService
}

func NewHealthHandler(s *service.HealthService) *HealthHandler {
	return &HealthHandler{service: s}
}

func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": h.service.GetPingMessage()})
}
