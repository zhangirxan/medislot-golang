package service

import (
	"log/slog"

	"medislot/internal/models"
	"medislot/internal/repository"
)

type SlotService interface {
	CreateSlot(doctorID string, req *models.CreateSlotRequest) (*models.TimeSlot, error)
	GetAvailable() ([]*models.TimeSlot, error)
	GetMySlots(doctorID string) ([]*models.TimeSlot, error)
	CancelSlot(doctorID, slotID string) error
}

type slotService struct {
	slotRepo repository.SlotRepository
}

func NewSlotService(slotRepo repository.SlotRepository) SlotService {
	return &slotService{slotRepo: slotRepo}
}

func (s *slotService) CreateSlot(doctorID string, req *models.CreateSlotRequest) (*models.TimeSlot, error) {
	
	if err := req.Validate(); err != nil {
		slog.Warn("slot validation failed", "doctor_id", doctorID, "error", err)
		return nil, err
	}

	overlap, err := s.slotRepo.HasOverlap(doctorID, req.StartTime, req.EndTime, "")
	if err != nil {
		slog.Error("overlap check failed", "doctor_id", doctorID, "error", err)
		return nil, err
	}
	if overlap {
		slog.Warn("slot overlap detected", "doctor_id", doctorID, "start", req.StartTime)
		return nil, models.ErrSlotOverlap
	}

	slot := &models.TimeSlot{
		DoctorID:  doctorID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	if err := s.slotRepo.Create(slot); err != nil {
		slog.Error("slot creation failed", "doctor_id", doctorID, "error", err)
		return nil, err
	}

	slog.Info("slot created", "slot_id", slot.ID, "doctor_id", doctorID)
	return slot, nil
}

func (s *slotService) GetAvailable() ([]*models.TimeSlot, error) {
	return s.slotRepo.GetAvailable()
}

func (s *slotService) GetMySlots(doctorID string) ([]*models.TimeSlot, error) {
	return s.slotRepo.GetByDoctor(doctorID)
}

func (s *slotService) CancelSlot(doctorID, slotID string) error {
	slot, err := s.slotRepo.GetByID(slotID)
	if err != nil {
		return err
	}
	if slot.DoctorID != doctorID {
		slog.Warn("unauthorised slot cancel attempt",
			"slot_id", slotID,
			"owner", slot.DoctorID,
			"caller", doctorID,
		)
		return models.ErrForbidden
	}
	if slot.Status != models.SlotAvailable {
		return models.ErrSlotNotAvailable
	}

	if err := s.slotRepo.UpdateStatus(slotID, models.SlotCancelled); err != nil {
		slog.Error("slot status update failed", "slot_id", slotID, "error", err)
		return err
	}

	slog.Info("slot cancelled", "slot_id", slotID, "doctor_id", doctorID)
	return nil
}
