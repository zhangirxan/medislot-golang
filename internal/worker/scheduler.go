package worker

import (
	"context"
	"log/slog"
	"time"

	"medislot/internal/models"
	"medislot/internal/repository"
)

type Scheduler struct {
	apptRepo        repository.AppointmentRepository
	interval        time.Duration
	expiryThreshold time.Duration
	reminderChan    chan string 
}

func NewScheduler(
	apptRepo repository.AppointmentRepository,
	intervalSeconds int,
	expiryMinutes int,
) *Scheduler {
	return &Scheduler{
		apptRepo:        apptRepo,
		interval:        time.Duration(intervalSeconds) * time.Second,
		expiryThreshold: time.Duration(expiryMinutes) * time.Minute,
		reminderChan:    make(chan string, 100),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("background scheduler started",
		"interval", s.interval.String(),
		"expiry_threshold", s.expiryThreshold.String(),
	)

	go s.runTicker(ctx)

	go s.processReminders(ctx)

	<-ctx.Done()
	slog.Info("background scheduler stopped")
}

func (s *Scheduler) runTicker(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cancelExpiredAppointments()
			s.enqueueReminders()
		}
	}
}

func (s *Scheduler) cancelExpiredAppointments() {
	cutoff := time.Now().Add(-s.expiryThreshold)

	appts, err := s.apptRepo.GetExpiredPending(cutoff)
	if err != nil {
		slog.Error("worker: GetExpiredPending failed", "error", err)
		return
	}
	if len(appts) == 0 {
		slog.Debug("worker: no expired appointments found")
		return
	}

	ids := make([]string, len(appts))
	for i, a := range appts {
		ids[i] = a.ID
	}

	if err := s.apptRepo.BulkUpdateStatus(ids, models.AppointmentExpired); err != nil {
		slog.Error("worker: BulkUpdateStatus failed", "error", err, "count", len(ids))
		return
	}

	slog.Info("worker: expired appointments processed", "count", len(ids))
}

func (s *Scheduler) enqueueReminders() {
	select {
	case s.reminderChan <- "heartbeat":
		slog.Debug("worker: reminder enqueued")
	default:
		slog.Warn("worker: reminder channel full, skipping tick")
	}
}

func (s *Scheduler) processReminders(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case id := <-s.reminderChan:
			slog.Info("worker: reminder dispatched", "target", id)
		}
	}
}
