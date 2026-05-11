package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"medislot/internal/service"
	"medislot/pkg/utils"
)

type StatsHandler struct{ statsSvc service.StatsService }

func NewStatsHandler(statsSvc service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: statsSvc}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	stats, err := h.statsSvc.GetSystemStats()
	if err != nil {
		slog.Error("get stats failed", "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, stats)
}
