package models

type UserStats struct {
	TotalUsers    int `json:"total_users"`
	TotalAdmins   int `json:"total_admins"`
	TotalDoctors  int `json:"total_doctors"`
	TotalPatients int `json:"total_patients"`
}

type SlotStats struct {
	TotalSlots int `json:"total_slots"`
	Available  int `json:"available"`
	Booked     int `json:"booked"`
	Cancelled  int `json:"cancelled"`
}

type AppointmentStats struct {
	TotalAppointments int `json:"total_appointments"`
	Pending           int `json:"pending"`
	Confirmed         int `json:"confirmed"`
	Cancelled         int `json:"cancelled"`
	Expired           int `json:"expired"`
}

type TopDoctor struct {
	DoctorID   string `json:"doctor_id"`
	DoctorName string `json:"doctor_name"`
	Bookings   int    `json:"bookings"`
}

type BusiestDay struct {
	DayOfWeek string `json:"day_of_week"`
	Bookings  int    `json:"bookings"`
}

type SystemStats struct {
	Users        UserStats        `json:"users"`
	Slots        SlotStats        `json:"slots"`
	Appointments AppointmentStats `json:"appointments"`
	TopDoctors   []TopDoctor      `json:"top_doctors"`
	BusiestDays  []BusiestDay     `json:"busiest_days"`
}
