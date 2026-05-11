package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"medislot/internal/models"
	"medislot/internal/service"
	"medislot/pkg/middleware"
	"medislot/pkg/utils"
)

type SlotHandler struct{ slotSvc service.SlotService }

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
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "could not identify caller"})
		return
	}
	slot, err := h.slotSvc.CreateSlot(doctorID, &req)
	if err != nil {
		slog.Warn("create slot failed", "doctor_id", doctorID, "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	slog.Info("slot created", "slot_id", slot.ID, "doctor_id", doctorID)
	utils.SuccessResponse(c, http.StatusCreated, slot)
}

func (h *SlotHandler) BulkCreate(c *gin.Context) {
	var req models.BulkCreateSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}
	doctorID := middleware.GetCallerID(c)
	if doctorID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "could not identify caller"})
		return
	}
	result, err := h.slotSvc.BulkCreateSlots(doctorID, &req)
	if err != nil {
		slog.Warn("bulk create slots failed", "doctor_id", doctorID, "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	slog.Info("bulk slots created", "doctor_id", doctorID, "total_created", result.TotalCreated, "total_skipped", result.TotalSkipped)
	utils.SuccessResponse(c, http.StatusCreated, result)
}

func (h *SlotHandler) GetAvailable(c *gin.Context) {
	filter, err := parseSlotFilter(c)
	if err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}
	slots, err := h.slotSvc.GetAvailableFiltered(filter)
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
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "could not identify caller"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "could not identify caller"})
		return
	}
	if err := h.slotSvc.CancelSlot(doctorID, slotID); err != nil {
		slog.Warn("cancel slot failed", "slot_id", slotID, "doctor_id", doctorID, "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	slog.Info("slot cancelled", "slot_id", slotID, "doctor_id", doctorID)
	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "slot cancelled"})
}

func parseSlotFilter(c *gin.Context) (models.SlotFilter, error) {
	var f models.SlotFilter

	if raw := c.Query("doctor_type"); raw != "" {
		dt := models.DoctorType(raw)
		if !models.IsValidDoctorType(dt) {
			return f, models.ErrInvalidDoctorType
		}
		f.DoctorType = &dt
	}
	if raw := c.Query("month"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 12 {
			return f, models.ErrInvalidFilterMonth
		}
		f.Month = &v
	}
	if raw := c.Query("day"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 31 {
			return f, models.ErrInvalidFilterDay
		}
		f.Day = &v
	}
	if raw := c.Query("hour_from"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 || v > 23 {
			return f, models.ErrInvalidFilterHour
		}
		f.HourFrom = &v
	}
	if raw := c.Query("hour_to"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 || v > 23 {
			return f, models.ErrInvalidFilterHour
		}
		f.HourTo = &v
	}
	return f, nil
}
