package repository

import (
	"database/sql"

	"medislot/internal/models"
)

type StatsRepository interface {
	GetSystemStats() (*models.SystemStats, error)
}

type statsRepository struct{ db *sql.DB }

func NewStatsRepository(db *sql.DB) StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) GetSystemStats() (*models.SystemStats, error) {
	stats := &models.SystemStats{}

	if err := r.db.QueryRow(`
		SELECT
			COUNT(*)                                          AS total,
			COUNT(*) FILTER (WHERE role = 'admin')           AS admins,
			COUNT(*) FILTER (WHERE role = 'doctor')          AS doctors,
			COUNT(*) FILTER (WHERE role = 'patient')         AS patients
		FROM users`).Scan(
		&stats.Users.TotalUsers,
		&stats.Users.TotalAdmins,
		&stats.Users.TotalDoctors,
		&stats.Users.TotalPatients,
	); err != nil {
		return nil, err
	}

	if err := r.db.QueryRow(`
		SELECT
			COUNT(*)                                              AS total,
			COUNT(*) FILTER (WHERE status = 'available')         AS available,
			COUNT(*) FILTER (WHERE status = 'booked')            AS booked,
			COUNT(*) FILTER (WHERE status = 'cancelled')         AS cancelled
		FROM time_slots`).Scan(
		&stats.Slots.TotalSlots,
		&stats.Slots.Available,
		&stats.Slots.Booked,
		&stats.Slots.Cancelled,
	); err != nil {
		return nil, err
	}

	if err := r.db.QueryRow(`
		SELECT
			COUNT(*)                                               AS total,
			COUNT(*) FILTER (WHERE status = 'pending')            AS pending,
			COUNT(*) FILTER (WHERE status = 'confirmed')          AS confirmed,
			COUNT(*) FILTER (WHERE status = 'cancelled')          AS cancelled,
			COUNT(*) FILTER (WHERE status = 'expired')            AS expired
		FROM appointments`).Scan(
		&stats.Appointments.TotalAppointments,
		&stats.Appointments.Pending,
		&stats.Appointments.Confirmed,
		&stats.Appointments.Cancelled,
		&stats.Appointments.Expired,
	); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`
		SELECT u.id, u.name, COUNT(a.id) AS bookings
		FROM appointments a
		JOIN time_slots s ON s.id = a.slot_id
		JOIN users      u ON u.id = s.doctor_id
		WHERE a.status != 'cancelled'
		GROUP BY u.id, u.name
		ORDER BY bookings DESC
		LIMIT 5`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var td models.TopDoctor
		if err := rows.Scan(&td.DoctorID, &td.DoctorName, &td.Bookings); err != nil {
			return nil, err
		}
		stats.TopDoctors = append(stats.TopDoctors, td)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	dayRows, err := r.db.Query(`
		SELECT TO_CHAR(s.start_time, 'Day') AS day_name,
		       COUNT(a.id)                  AS bookings
		FROM appointments a
		JOIN time_slots s ON s.id = a.slot_id
		WHERE a.status != 'cancelled'
		GROUP BY day_name
		ORDER BY bookings DESC`)
	if err != nil {
		return nil, err
	}
	defer dayRows.Close()
	for dayRows.Next() {
		var bd models.BusiestDay
		if err := dayRows.Scan(&bd.DayOfWeek, &bd.Bookings); err != nil {
			return nil, err
		}
		stats.BusiestDays = append(stats.BusiestDays, bd)
	}

	return stats, dayRows.Err()
}
