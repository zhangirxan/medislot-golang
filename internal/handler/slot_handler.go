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

type SlotHandler struct {
	slotSvc service.SlotService
}

func NewSlotHandler(slotSvc service.SlotService) *SlotHandler {
	return &SlotHandler{slotSvc: slotSvc}
}

func (h *SlotHandler) Create(c *gin.Context) {
	var req models.CreateSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	doctorID := middleware.GetCallerID(c)
	if doctorID == "" {
		slog.Error("create slot: caller ID missing from context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "could not identify caller — ensure Authorization: Bearer <token> header is set",
		})
		return
	}

	slot, err := h.slotSvc.CreateSlot(doctorID, &req)
	if err != nil {
		slog.Warn("create slot failed",
			"doctor_id", doctorID,
			"start", req.StartTime,
			"error", err,
		)
		utils.ErrorResponse(c, err)
		return
	}

	slog.Info("slot created", "slot_id", slot.ID, "doctor_id", doctorID)
	utils.SuccessResponse(c, http.StatusCreated, slot)
}

func (h *SlotHandler) GetAvailable(c *gin.Context) {
	slots, err := h.slotSvc.GetAvailable()
	if err != nil {
		slog.Error("get available slots failed", "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, slots)
}

func (h *SlotHandler) GetMySlots(c *gin.Context) {
	doctorID := middleware.GetCallerID(c)
	if doctorID == "" {
		slog.Error("get my slots: caller ID missing from context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "could not identify caller — ensure Authorization: Bearer <token> header is set",
		})
		return
	}

	slots, err := h.slotSvc.GetMySlots(doctorID)
	if err != nil {
		slog.Error("get my slots failed", "doctor_id", doctorID, "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, slots)
}

func (h *SlotHandler) Cancel(c *gin.Context) {
	slotID := c.Param("id")
	doctorID := middleware.GetCallerID(c)
	if doctorID == "" {
		slog.Error("cancel slot: caller ID missing from context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "could not identify caller — ensure Authorization: Bearer <token> header is set",
		})
		return
	}

	if err := h.slotSvc.CancelSlot(doctorID, slotID); err != nil {
		slog.Warn("cancel slot failed",
			"slot_id", slotID,
			"doctor_id", doctorID,
			"error", err,
		)
		utils.ErrorResponse(c, err)
		return
	}

	slog.Info("slot cancelled", "slot_id", slotID, "doctor_id", doctorID)
	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "slot cancelled"})
}
