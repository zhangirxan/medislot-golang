package models

import "time"

type AppointmentStatus string

const (
	AppointmentPending   AppointmentStatus = "pending"
	AppointmentConfirmed AppointmentStatus = "confirmed"
	AppointmentCancelled AppointmentStatus = "cancelled"
	AppointmentExpired   AppointmentStatus = "expired"
)

type Appointment struct {
	ID         string            `json:"id"          db:"id"`
	PatientID  string            `json:"patient_id"  db:"patient_id"`
	SlotID     string            `json:"slot_id"     db:"slot_id"`
	Status     AppointmentStatus `json:"status"      db:"status"`
	Notes      string            `json:"notes"       db:"notes"`
	CreatedAt  time.Time         `json:"created_at"  db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"  db:"updated_at"`

	// Populated via joins.
	PatientName string     `json:"patient_name,omitempty" db:"patient_name"`
	DoctorName  string     `json:"doctor_name,omitempty"  db:"doctor_name"`
	StartTime   *time.Time `json:"start_time,omitempty"   db:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"     db:"end_time"`
}


type BookAppointmentRequest struct {
	SlotID string `json:"slot_id" binding:"required,uuid"`
	Notes  string `json:"notes"   binding:"omitempty,max=500"`
}
