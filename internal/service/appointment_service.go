package service

import (
	"log/slog"

	"medislot/internal/models"
	"medislot/internal/repository"
)

type AppointmentService interface {
	Book(patientID string, req *models.BookAppointmentRequest) (*models.Appointment, error)
	GetMyAppointments(patientID string) ([]*models.Appointment, error)
	GetAll() ([]*models.Appointment, error)
	GetByID(id string) (*models.Appointment, error)
	Cancel(callerID string, callerRole models.Role, appointmentID string) error
}

type appointmentService struct {
	apptRepo repository.AppointmentRepository
}

func NewAppointmentService(apptRepo repository.AppointmentRepository) AppointmentService {
	return &appointmentService{apptRepo: apptRepo}
}

func (s *appointmentService) Book(patientID string, req *models.BookAppointmentRequest) (*models.Appointment, error) {
	appt, err := s.apptRepo.BookWithTx(patientID, req.SlotID, req.Notes)
	if err != nil {
		slog.Warn("booking failed",
			"patient_id", patientID,
			"slot_id", req.SlotID,
			"error", err,
		)
		return nil, err
	}
	slog.Info("appointment booked", "appointment_id", appt.ID, "patient_id", patientID)
	return appt, nil
}

func (s *appointmentService) GetMyAppointments(patientID string) ([]*models.Appointment, error) {
	return s.apptRepo.GetByPatient(patientID)
}

func (s *appointmentService) GetAll() ([]*models.Appointment, error) {
	return s.apptRepo.GetAll()
}

func (s *appointmentService) GetByID(id string) (*models.Appointment, error) {
	return s.apptRepo.GetByID(id)
}

func (s *appointmentService) Cancel(callerID string, callerRole models.Role, appointmentID string) error {
	appt, err := s.apptRepo.GetByID(appointmentID)
	if err != nil {
		return err
	}

	if callerRole != models.RoleAdmin && appt.PatientID != callerID {
		slog.Warn("unauthorised cancel attempt",
			"appointment_id", appointmentID,
			"owner", appt.PatientID,
			"caller", callerID,
		)
		return models.ErrNotAppointmentOwner
	}

	if err := s.apptRepo.Cancel(appointmentID); err != nil {
		slog.Error("appointment cancel failed", "appointment_id", appointmentID, "error", err)
		return err
	}

	slog.Info("appointment cancelled", "appointment_id", appointmentID, "by", callerID)
	return nil
}
