package models

import "time"

type SlotStatus string

const (
	SlotAvailable  SlotStatus = "available"
	SlotBooked     SlotStatus = "booked"
	SlotCancelled  SlotStatus = "cancelled"
)

type TimeSlot struct {
	ID        string     `json:"id"         db:"id"`
	DoctorID  string     `json:"doctor_id"  db:"doctor_id"`
	StartTime time.Time  `json:"start_time" db:"start_time"`
	EndTime   time.Time  `json:"end_time"   db:"end_time"`
	Status    SlotStatus `json:"status"     db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`

	DoctorName string `json:"doctor_name,omitempty" db:"doctor_name"`
}


type CreateSlotRequest struct {
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time"   binding:"required"`
}

func (r *CreateSlotRequest) Validate() error {
	if !r.EndTime.After(r.StartTime) {
		return ErrSlotEndBeforeStart
	}
	if r.StartTime.Before(time.Now()) {
		return ErrSlotInPast
	}
	return nil
}
