package models

import "errors"


var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden: insufficient permissions")
)


var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyTaken = errors.New("email is already registered")
)


var (
	ErrSlotNotFound      = errors.New("time slot not found")
	ErrSlotNotAvailable  = errors.New("time slot is not available")
	ErrSlotOverlap       = errors.New("time slot overlaps with an existing slot")
	ErrSlotEndBeforeStart = errors.New("end_time must be after start_time")
	ErrSlotInPast        = errors.New("start_time must be in the future")
)


var (
	ErrAppointmentNotFound    = errors.New("appointment not found")
	ErrAlreadyBooked          = errors.New("slot is already booked")
	ErrNotAppointmentOwner    = errors.New("you are not the owner of this appointment")
	ErrAppointmentNotCancellable = errors.New("appointment cannot be cancelled in its current state")
)
