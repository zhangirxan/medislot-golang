package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"medislot/internal/models"
)

type AppointmentRepository interface {
	BookWithTx(patientID, slotID, notes string) (*models.Appointment, error)
	GetByID(id string) (*models.Appointment, error)
	GetByPatient(patientID string) ([]*models.Appointment, error)
	GetByDoctor(doctorID string) ([]*models.Appointment, error)
	GetAll() ([]*models.Appointment, error)
	Confirm(id, doctorID string) error
	Cancel(id string) error
	GetExpiredPending(before time.Time) ([]*models.Appointment, error)
	ExpirePending(ids []string) error
}

type appointmentRepository struct{ db *sql.DB }

func NewAppointmentRepository(db *sql.DB) AppointmentRepository {
	return &appointmentRepository{db: db}
}

func (r *appointmentRepository) BookWithTx(patientID, slotID, notes string) (*models.Appointment, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var status models.SlotStatus
	err = tx.QueryRow(`SELECT status FROM time_slots WHERE id=$1 FOR UPDATE`, slotID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrSlotNotFound
	}
	if err != nil {
		return nil, err
	}
	if status != models.SlotAvailable {
		return nil, models.ErrAlreadyBooked
	}

	if _, err = tx.Exec(`UPDATE time_slots SET status='booked', updated_at=$1 WHERE id=$2`,
		time.Now(), slotID); err != nil {
		return nil, err
	}

	appt := &models.Appointment{
		ID:        uuid.NewString(),
		PatientID: patientID,
		SlotID:    slotID,
		Status:    models.AppointmentPending,
		Notes:     notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if _, err = tx.Exec(`
		INSERT INTO appointments (id, patient_id, slot_id, status, notes, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		appt.ID, appt.PatientID, appt.SlotID,
		appt.Status, appt.Notes, appt.CreatedAt, appt.UpdatedAt,
	); err != nil {
		if isUniqueViolation(err) {
			return nil, models.ErrAlreadyBooked
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return appt, nil
}

func (r *appointmentRepository) GetByID(id string) (*models.Appointment, error) {
	a := &models.Appointment{}
	err := r.db.QueryRow(`
		SELECT a.id, a.patient_id, s.doctor_id, a.slot_id, a.status, a.notes,
		       a.created_at, a.updated_at,
		       p.name, p.medical_notes,
		       d.name, s.start_time, s.end_time
		FROM appointments a
		JOIN users      p ON p.id = a.patient_id
		JOIN time_slots s ON s.id = a.slot_id
		JOIN users      d ON d.id = s.doctor_id
		WHERE a.id = $1`, id).Scan(
		&a.ID, &a.PatientID, &a.DoctorID, &a.SlotID, &a.Status, &a.Notes,
		&a.CreatedAt, &a.UpdatedAt,
		&a.PatientName, &a.PatientMedicalNotes,
		&a.DoctorName, &a.StartTime, &a.EndTime,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrAppointmentNotFound
	}
	return a, err
}

func (r *appointmentRepository) GetByPatient(patientID string) ([]*models.Appointment, error) {
	return r.scanAppointments(r.db.Query(`
		SELECT a.id, a.patient_id, s.doctor_id, a.slot_id, a.status, a.notes,
		       a.created_at, a.updated_at,
		       p.name, p.medical_notes,
		       d.name, s.start_time, s.end_time
		FROM appointments a
		JOIN users      p ON p.id = a.patient_id
		JOIN time_slots s ON s.id = a.slot_id
		JOIN users      d ON d.id = s.doctor_id
		WHERE a.patient_id = $1
		ORDER BY a.created_at DESC`, patientID))
}

func (r *appointmentRepository) GetByDoctor(doctorID string) ([]*models.Appointment, error) {
	return r.scanAppointments(r.db.Query(`
		SELECT a.id, a.patient_id, s.doctor_id, a.slot_id, a.status, a.notes,
		       a.created_at, a.updated_at,
		       p.name, p.medical_notes,
		       d.name, s.start_time, s.end_time
		FROM appointments a
		JOIN users      p ON p.id = a.patient_id
		JOIN time_slots s ON s.id = a.slot_id
		JOIN users      d ON d.id = s.doctor_id
		WHERE s.doctor_id = $1
		ORDER BY a.created_at DESC`, doctorID))
}

func (r *appointmentRepository) GetAll() ([]*models.Appointment, error) {
	return r.scanAppointments(r.db.Query(`
		SELECT a.id, a.patient_id, s.doctor_id, a.slot_id, a.status, a.notes,
		       a.created_at, a.updated_at,
		       p.name, p.medical_notes,
		       d.name, s.start_time, s.end_time
		FROM appointments a
		JOIN users      p ON p.id = a.patient_id
		JOIN time_slots s ON s.id = a.slot_id
		JOIN users      d ON d.id = s.doctor_id
		ORDER BY a.created_at DESC`))
}

func (r *appointmentRepository) Confirm(id, doctorID string) error {
	res, err := r.db.Exec(`
		UPDATE appointments a
		SET    status = 'confirmed', updated_at = $1
		FROM   time_slots s
		WHERE  a.slot_id   = s.id
		  AND  a.id        = $2
		  AND  s.doctor_id = $3
		  AND  a.status    = 'pending'`,
		time.Now(), id, doctorID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return models.ErrAppointmentNotConfirmable
	}
	return nil
}

func (r *appointmentRepository) Cancel(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var slotID string
	var status models.AppointmentStatus
	err = tx.QueryRow(`SELECT slot_id, status FROM appointments WHERE id=$1 FOR UPDATE`, id).
		Scan(&slotID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrAppointmentNotFound
	}
	if err != nil {
		return err
	}
	if status == models.AppointmentCancelled || status == models.AppointmentExpired {
		return models.ErrAppointmentNotCancellable
	}

	if _, err = tx.Exec(`UPDATE appointments SET status='cancelled', updated_at=$1 WHERE id=$2`,
		time.Now(), id); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE time_slots SET status='available', updated_at=$1 WHERE id=$2`,
		time.Now(), slotID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *appointmentRepository) GetExpiredPending(before time.Time) ([]*models.Appointment, error) {
	return r.scanAppointments(r.db.Query(`
		SELECT a.id, a.patient_id, s.doctor_id, a.slot_id, a.status, a.notes,
		       a.created_at, a.updated_at,
		       p.name, p.medical_notes,
		       d.name, s.start_time, s.end_time
		FROM appointments a
		JOIN users      p ON p.id = a.patient_id
		JOIN time_slots s ON s.id = a.slot_id
		JOIN users      d ON d.id = s.doctor_id
		WHERE a.status='pending' AND a.created_at<$1`, before))
}

func (r *appointmentRepository) ExpirePending(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	args := make([]interface{}, len(ids))
	placeholders := ""
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "$" + itoa(i+1)
		args[i] = id
	}
	_, err := r.db.Exec(`
		WITH expired AS (
			UPDATE appointments
			SET status='expired', updated_at=NOW()
			WHERE id IN (`+placeholders+`) AND status='pending'
			RETURNING slot_id
		)
		UPDATE time_slots
		SET status='available', updated_at=NOW()
		WHERE id IN (SELECT slot_id FROM expired)
		  AND status='booked'`,
		args...,
	)
	return err
}

func (r *appointmentRepository) scanAppointments(rows *sql.Rows, err error) ([]*models.Appointment, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*models.Appointment
	for rows.Next() {
		a := &models.Appointment{}
		if err := rows.Scan(
			&a.ID, &a.PatientID, &a.DoctorID, &a.SlotID, &a.Status, &a.Notes,
			&a.CreatedAt, &a.UpdatedAt,
			&a.PatientName, &a.PatientMedicalNotes,
			&a.DoctorName, &a.StartTime, &a.EndTime,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
