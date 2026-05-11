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
	ErrSlotNotFound         = errors.New("time slot not found")
	ErrSlotNotAvailable     = errors.New("time slot is not available")
	ErrSlotOverlap          = errors.New("time slot overlaps with an existing slot")
	ErrSlotEndBeforeStart   = errors.New("end_time must be after start_time")
	ErrSlotInPast           = errors.New("start_time must be in the future")
	ErrSlotDurationTooShort = errors.New("slot_duration_minutes must be at least 5 minutes")
	ErrTimeFrameTooShort    = errors.New("time frame is shorter than one slot duration")
	ErrNoSlotsCreated       = errors.New("no slots were created - all windows overlap with existing slots")
)

var (
	ErrAppointmentNotFound       = errors.New("appointment not found")
	ErrAlreadyBooked             = errors.New("slot is already booked")
	ErrNotAppointmentOwner       = errors.New("you are not the owner of this appointment")
	ErrAppointmentNotCancellable = errors.New("appointment cannot be cancelled in its current state")
)

var (
	ErrInvalidDoctorType       = errors.New("invalid doctor_type - valid values: therapist, gynecologist, psychotherapist, cardiologist, neurologist, dermatologist, pediatrician, surgeon, ophthalmologist, orthopedist")
	ErrDoctorTypeRequired      = errors.New("doctor_type is required when role is doctor")
	ErrDoctorTypeOnlyForDoctor = errors.New("doctor_type can only be set when role is doctor")
	ErrInvalidFilterMonth      = errors.New("invalid month - must be an integer between 1 and 12")
	ErrInvalidFilterDay        = errors.New("invalid day - must be an integer between 1 and 31")
	ErrInvalidFilterHour       = errors.New("invalid hour - must be an integer between 0 and 23")
)

var (
	ErrAppointmentNotConfirmable = errors.New("appointment cannot be confirmed - it may not be pending, not yours, or not found")
)
