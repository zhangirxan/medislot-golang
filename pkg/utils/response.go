package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"medislot/internal/models"
)

func SuccessResponse(c *gin.Context, code int, data interface{}) {
	c.JSON(code, gin.H{
		"success": true,
		"data":    data,
	})
}

func ErrorResponse(c *gin.Context, err error) {
	code := httpStatusFromError(err)
	c.JSON(code, gin.H{
		"success": false,
		"error":   err.Error(),
	})
}

func ValidationErrorResponse(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error":   "validation failed: " + err.Error(),
	})
}

func httpStatusFromError(err error) int {
	switch {
	case errors.Is(err, models.ErrUnauthorized),
		errors.Is(err, models.ErrInvalidCredentials):
		return http.StatusUnauthorized

	case errors.Is(err, models.ErrForbidden),
		errors.Is(err, models.ErrNotAppointmentOwner):
		return http.StatusForbidden

	case errors.Is(err, models.ErrUserNotFound),
		errors.Is(err, models.ErrSlotNotFound),
		errors.Is(err, models.ErrAppointmentNotFound):
		return http.StatusNotFound

	case errors.Is(err, models.ErrEmailAlreadyTaken),
		errors.Is(err, models.ErrSlotOverlap),
		errors.Is(err, models.ErrAlreadyBooked),
		errors.Is(err, models.ErrSlotNotAvailable),
		errors.Is(err, models.ErrSlotEndBeforeStart),
		errors.Is(err, models.ErrSlotInPast),
		errors.Is(err, models.ErrAppointmentNotCancellable):
		return http.StatusConflict

	default:
		return http.StatusInternalServerError
	}
}
