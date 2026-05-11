package service

import (
	"log/slog"
	"time"

	"medislot/internal/models"
	"medislot/internal/repository"
)

type SlotService interface {
	CreateSlot(doctorID string, req *models.CreateSlotRequest) (*models.TimeSlot, error)
	BulkCreateSlots(doctorID string, req *models.BulkCreateSlotRequest) (*models.BulkCreateResult, error)
	GetAvailableFiltered(f models.SlotFilter) ([]*models.TimeSlot, error)
	GetMySlots(doctorID string) ([]*models.TimeSlot, error)
	CancelSlot(doctorID, slotID string) error
}

type slotService struct{ slotRepo repository.SlotRepository }

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
		slog.Warn("slot overlap detected", "doctor_id", doctorID)
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

func (s *slotService) BulkCreateSlots(doctorID string, req *models.BulkCreateSlotRequest) (*models.BulkCreateResult, error) {
	if err := req.Validate(); err != nil {
		slog.Warn("bulk slot validation failed", "doctor_id", doctorID, "error", err)
		return nil, err
	}

	duration := time.Duration(req.SlotDurationMinutes) * time.Minute

	var toCreate []*models.TimeSlot
	var skipped []models.SkippedSlot

	cursor := req.WorkStart
	for {
		slotEnd := cursor.Add(duration)
		if slotEnd.After(req.WorkEnd) {
			break
		}

		overlap, err := s.slotRepo.HasOverlap(doctorID, cursor, slotEnd, "")
		if err != nil {
			slog.Error("overlap check failed during bulk create", "doctor_id", doctorID, "start", cursor, "error", err)
			return nil, err
		}

		if overlap {
			skipped = append(skipped, models.SkippedSlot{
				StartTime: cursor,
				EndTime:   slotEnd,
				Reason:    "overlaps with an existing slot",
			})
			slog.Warn("bulk: slot window skipped", "doctor_id", doctorID, "start", cursor, "end", slotEnd)
		} else {
			toCreate = append(toCreate, &models.TimeSlot{
				DoctorID:  doctorID,
				StartTime: cursor,
				EndTime:   slotEnd,
			})
		}
		cursor = slotEnd
	}

	if len(toCreate) == 0 {
		if len(skipped) == 0 {
			slog.Warn("bulk: no slots to create", "doctor_id", doctorID)
			return nil, models.ErrNoSlotsCreated
		}
		slog.Info("bulk slots skipped", "doctor_id", doctorID, "skipped", len(skipped))
		return &models.BulkCreateResult{
			Created:      []*models.TimeSlot{},
			Skipped:      skipped,
			TotalCreated: 0,
			TotalSkipped: len(skipped),
		}, nil
	}

	if err := s.slotRepo.BulkCreate(toCreate); err != nil {
		slog.Error("bulk create failed", "doctor_id", doctorID, "error", err)
		return nil, err
	}

	slog.Info("bulk slots created", "doctor_id", doctorID, "created", len(toCreate), "skipped", len(skipped))
	return &models.BulkCreateResult{
		Created:      toCreate,
		Skipped:      skipped,
		TotalCreated: len(toCreate),
		TotalSkipped: len(skipped),
	}, nil
}

func (s *slotService) GetAvailableFiltered(f models.SlotFilter) ([]*models.TimeSlot, error) {
	return s.slotRepo.GetAvailableFiltered(f)
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
		slog.Warn("unauthorised slot cancel", "slot_id", slotID, "owner", slot.DoctorID, "caller", doctorID)
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
