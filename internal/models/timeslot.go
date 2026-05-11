package models

import "time"

type SlotStatus string

const (
	SlotAvailable SlotStatus = "available"
	SlotBooked    SlotStatus = "booked"
	SlotCancelled SlotStatus = "cancelled"
)

type TimeSlot struct {
	ID         string      `json:"id"                    db:"id"`
	DoctorID   string      `json:"doctor_id"             db:"doctor_id"`
	StartTime  time.Time   `json:"start_time"            db:"start_time"`
	EndTime    time.Time   `json:"end_time"              db:"end_time"`
	Status     SlotStatus  `json:"status"                db:"status"`
	CreatedAt  time.Time   `json:"created_at"            db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"            db:"updated_at"`
	DoctorName string      `json:"doctor_name,omitempty" db:"doctor_name"`
	DoctorType *DoctorType `json:"doctor_type,omitempty" db:"doctor_type"`
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

type BulkCreateSlotRequest struct {
	WorkStart           time.Time `json:"work_start"            binding:"required"`
	WorkEnd             time.Time `json:"work_end"              binding:"required"`
	SlotDurationMinutes int       `json:"slot_duration_minutes" binding:"required,min=5,max=480"`
}

func (r *BulkCreateSlotRequest) Validate() error {
	if !r.WorkEnd.After(r.WorkStart) {
		return ErrSlotEndBeforeStart
	}
	if r.WorkStart.Before(time.Now()) {
		return ErrSlotInPast
	}
	if r.SlotDurationMinutes < 5 {
		return ErrSlotDurationTooShort
	}
	totalMinutes := int(r.WorkEnd.Sub(r.WorkStart).Minutes())
	if totalMinutes < r.SlotDurationMinutes {
		return ErrTimeFrameTooShort
	}
	return nil
}

type SkippedSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
}

type BulkCreateResult struct {
	Created      []*TimeSlot   `json:"created"`
	Skipped      []SkippedSlot `json:"skipped"`
	TotalCreated int           `json:"total_created"`
	TotalSkipped int           `json:"total_skipped"`
}
