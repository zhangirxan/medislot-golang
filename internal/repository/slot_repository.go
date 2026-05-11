package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"medislot/internal/models"
)

type SlotRepository interface {
	Create(slot *models.TimeSlot) error
	BulkCreate(slots []*models.TimeSlot) error
	GetByID(id string) (*models.TimeSlot, error)
	GetAvailableFiltered(f models.SlotFilter) ([]*models.TimeSlot, error)
	GetByDoctor(doctorID string) ([]*models.TimeSlot, error)
	HasOverlap(doctorID string, start, end time.Time, excludeID string) (bool, error)
	UpdateStatus(id string, status models.SlotStatus) error
	Delete(id string) error
}

type slotRepository struct{ db *sql.DB }

func NewSlotRepository(db *sql.DB) SlotRepository {
	return &slotRepository{db: db}
}

func (r *slotRepository) Create(slot *models.TimeSlot) error {
	slot.ID = uuid.NewString()
	slot.Status = models.SlotAvailable
	slot.CreatedAt = time.Now()
	slot.UpdatedAt = time.Now()

	_, err := r.db.Exec(`
		INSERT INTO time_slots (id, doctor_id, start_time, end_time, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		slot.ID, slot.DoctorID, slot.StartTime, slot.EndTime,
		slot.Status, slot.CreatedAt, slot.UpdatedAt,
	)
	return err
}

func (r *slotRepository) BulkCreate(slots []*models.TimeSlot) error {
	if len(slots) == 0 {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()
	for _, slot := range slots {
		slot.ID = uuid.NewString()
		slot.Status = models.SlotAvailable
		slot.CreatedAt = now
		slot.UpdatedAt = now

		if _, err = tx.Exec(`
			INSERT INTO time_slots (id, doctor_id, start_time, end_time, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			slot.ID, slot.DoctorID, slot.StartTime, slot.EndTime,
			slot.Status, slot.CreatedAt, slot.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *slotRepository) GetByID(id string) (*models.TimeSlot, error) {
	s := &models.TimeSlot{}
	err := r.db.QueryRow(`
		SELECT s.id, s.doctor_id, s.start_time, s.end_time, s.status,
		       s.created_at, s.updated_at, u.name, u.doctor_type
		FROM   time_slots s
		JOIN   users u ON u.id = s.doctor_id
		WHERE  s.id = $1`, id).Scan(
		&s.ID, &s.DoctorID, &s.StartTime, &s.EndTime, &s.Status,
		&s.CreatedAt, &s.UpdatedAt, &s.DoctorName, &s.DoctorType,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrSlotNotFound
	}
	return s, err
}

func (r *slotRepository) GetAvailableFiltered(f models.SlotFilter) ([]*models.TimeSlot, error) {
	base := `
		SELECT s.id, s.doctor_id, s.start_time, s.end_time, s.status,
		       s.created_at, s.updated_at, u.name, u.doctor_type
		FROM   time_slots s
		JOIN   users u ON u.id = s.doctor_id
		WHERE  s.status = 'available'
		  AND  s.start_time > NOW()`

	args := []interface{}{}
	idx := 1

	add := func(cond string, val interface{}) string {
		args = append(args, val)
		s := fmt.Sprintf(" AND %s$%d", cond, idx)
		idx++
		return s
	}

	if f.DoctorType != nil {
		base += add("u.doctor_type = ", *f.DoctorType)
	}
	if f.Month != nil {
		base += add("EXTRACT(MONTH FROM s.start_time) = ", *f.Month)
	}
	if f.Day != nil {
		base += add("EXTRACT(DAY FROM s.start_time) = ", *f.Day)
	}
	if f.HourFrom != nil {
		base += add("EXTRACT(HOUR FROM s.start_time) >= ", *f.HourFrom)
	}
	if f.HourTo != nil {
		base += add("EXTRACT(HOUR FROM s.start_time) < ", *f.HourTo)
	}

	base += " ORDER BY s.start_time ASC"
	return r.scanSlots(r.db.Query(base, args...))
}

func (r *slotRepository) GetByDoctor(doctorID string) ([]*models.TimeSlot, error) {
	return r.scanSlots(r.db.Query(`
		SELECT s.id, s.doctor_id, s.start_time, s.end_time, s.status,
		       s.created_at, s.updated_at, u.name, u.doctor_type
		FROM   time_slots s
		JOIN   users u ON u.id = s.doctor_id
		WHERE  s.doctor_id = $1
		ORDER  BY s.start_time ASC`, doctorID))
}

func (r *slotRepository) HasOverlap(doctorID string, start, end time.Time, excludeID string) (bool, error) {
	var query string
	var args []interface{}

	if excludeID == "" {
		query = `SELECT COUNT(*) FROM time_slots
			WHERE doctor_id=$1 AND status!='cancelled'
			  AND start_time<$3 AND end_time>$2`
		args = []interface{}{doctorID, start, end}
	} else {
		query = `SELECT COUNT(*) FROM time_slots
			WHERE doctor_id=$1 AND status!='cancelled'
			  AND id!=$2 AND start_time<$4 AND end_time>$3`
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
			&s.ID, &s.DoctorID, &s.StartTime, &s.EndTime, &s.Status,
			&s.CreatedAt, &s.UpdatedAt, &s.DoctorName, &s.DoctorType,
		); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}
