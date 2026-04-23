package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"medislot/internal/models"
)

type SlotRepository interface {
	Create(slot *models.TimeSlot) error
	GetByID(id string) (*models.TimeSlot, error)
	GetAvailable() ([]*models.TimeSlot, error)
	GetByDoctor(doctorID string) ([]*models.TimeSlot, error)
	HasOverlap(doctorID string, start, end time.Time, excludeID string) (bool, error)
	UpdateStatus(id string, status models.SlotStatus) error
	Delete(id string) error
}

type slotRepository struct {
	db *sql.DB
}

func NewSlotRepository(db *sql.DB) SlotRepository {
	return &slotRepository{db: db}
}

func (r *slotRepository) Create(slot *models.TimeSlot) error {
	slot.ID = uuid.NewString()
	slot.Status = models.SlotAvailable
	slot.CreatedAt = time.Now()
	slot.UpdatedAt = time.Now()

	query := `
		INSERT INTO time_slots (id, doctor_id, start_time, end_time, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(query,
		slot.ID, slot.DoctorID,
		slot.StartTime, slot.EndTime,
		slot.Status,
		slot.CreatedAt, slot.UpdatedAt,
	)
	return err
}

func (r *slotRepository) GetByID(id string) (*models.TimeSlot, error) {
	query := `
		SELECT s.id, s.doctor_id, s.start_time, s.end_time, s.status, s.created_at, s.updated_at,
		       u.name AS doctor_name
		FROM   time_slots s
		JOIN   users u ON u.id = s.doctor_id
		WHERE  s.id = $1`

	s := &models.TimeSlot{}
	err := r.db.QueryRow(query, id).Scan(
		&s.ID, &s.DoctorID, &s.StartTime, &s.EndTime,
		&s.Status, &s.CreatedAt, &s.UpdatedAt, &s.DoctorName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrSlotNotFound
	}
	return s, err
}

func (r *slotRepository) GetAvailable() ([]*models.TimeSlot, error) {
	query := `
		SELECT s.id, s.doctor_id, s.start_time, s.end_time, s.status, s.created_at, s.updated_at,
		       u.name AS doctor_name
		FROM   time_slots s
		JOIN   users u ON u.id = s.doctor_id
		WHERE  s.status = 'available' AND s.start_time > NOW()
		ORDER  BY s.start_time ASC`
	return r.scanSlots(r.db.Query(query))
}

func (r *slotRepository) GetByDoctor(doctorID string) ([]*models.TimeSlot, error) {
	query := `
		SELECT s.id, s.doctor_id, s.start_time, s.end_time, s.status, s.created_at, s.updated_at,
		       u.name AS doctor_name
		FROM   time_slots s
		JOIN   users u ON u.id = s.doctor_id
		WHERE  s.doctor_id = $1
		ORDER  BY s.start_time ASC`
	return r.scanSlots(r.db.Query(query, doctorID))
}

func (r *slotRepository) HasOverlap(doctorID string, start, end time.Time, excludeID string) (bool, error) {
	var (
		query string
		args  []interface{}
	)

	if excludeID == "" {
		query = `
			SELECT COUNT(*) FROM time_slots
			WHERE  doctor_id  = $1
			  AND  status     != 'cancelled'
			  AND  start_time <  $3
			  AND  end_time   >  $2`
		args = []interface{}{doctorID, start, end}
	} else {
		query = `
			SELECT COUNT(*) FROM time_slots
			WHERE  doctor_id  = $1
			  AND  status     != 'cancelled'
			  AND  id         != $2
			  AND  start_time <  $4
			  AND  end_time   >  $3`
		args = []interface{}{doctorID, excludeID, start, end}
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	return count > 0, err
}

func (r *slotRepository) UpdateStatus(id string, status models.SlotStatus) error {
	res, err := r.db.Exec(
		`UPDATE time_slots SET status=$1, updated_at=$2 WHERE id=$3`,
		status, time.Now(), id,
	)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, models.ErrSlotNotFound)
}

func (r *slotRepository) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM time_slots WHERE id=$1`, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, models.ErrSlotNotFound)
}


func (r *slotRepository) scanSlots(rows *sql.Rows, err error) ([]*models.TimeSlot, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []*models.TimeSlot
	for rows.Next() {
		s := &models.TimeSlot{}
		if err := rows.Scan(
			&s.ID, &s.DoctorID, &s.StartTime, &s.EndTime,
			&s.Status, &s.CreatedAt, &s.UpdatedAt, &s.DoctorName,
		); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}
