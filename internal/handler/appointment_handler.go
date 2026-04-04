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

type AppointmentHandler struct {
	apptSvc service.AppointmentService
}

func NewAppointmentHandler(apptSvc service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{apptSvc: apptSvc}
}

func (h *AppointmentHandler) Book(c *gin.Context) {
	var req models.BookAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	patientID := middleware.GetCallerID(c)
	appt, err := h.apptSvc.Book(patientID, &req)
	if err != nil {
		slog.Warn("book appointment failed",
			"patient_id", patientID,
			"slot_id", req.SlotID,
			"error", err,
		)
		utils.ErrorResponse(c, err)
		return
	}

	slog.Info("appointment booked",
		"appointment_id", appt.ID,
		"patient_id", patientID,
		"slot_id", req.SlotID,
	)
	utils.SuccessResponse(c, http.StatusCreated, appt)
}

func (h *AppointmentHandler) GetMyAppointments(c *gin.Context) {
	callerID := middleware.GetCallerID(c)
	callerRole := middleware.GetCallerRole(c)

	var (
		appts []*models.Appointment
		err   error
	)

	if callerRole == models.RoleAdmin {
		appts, err = h.apptSvc.GetAll()
	} else {
		appts, err = h.apptSvc.GetMyAppointments(callerID)
	}

	if err != nil {
		slog.Error("get appointments failed", "caller_id", callerID, "error", err)
		utils.ErrorResponse(c, err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, appts)
}

func (h *AppointmentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	appt, err := h.apptSvc.GetByID(id)
	if err != nil {
		slog.Warn("appointment not found", "id", id, "error", err)
		utils.ErrorResponse(c, err)
		return
	}

	callerID := middleware.GetCallerID(c)
	callerRole := middleware.GetCallerRole(c)
	if callerRole != models.RoleAdmin && appt.PatientID != callerID {
		utils.ErrorResponse(c, models.ErrForbidden)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, appt)
}

func (h *AppointmentHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	callerID := middleware.GetCallerID(c)
	callerRole := middleware.GetCallerRole(c)

	if err := h.apptSvc.Cancel(callerID, callerRole, id); err != nil {
		slog.Warn("cancel appointment failed",
			"appointment_id", id,
			"caller_id", callerID,
			"error", err,
		)
		utils.ErrorResponse(c, err)
		return
	}

	slog.Info("appointment cancelled", "appointment_id", id, "caller_id", callerID)
	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "appointment cancelled"})
}
