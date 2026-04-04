package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"medislot/internal/handler"
	"medislot/internal/repository"
	"medislot/internal/service"
	"medislot/internal/worker"
	"medislot/pkg/config"
	"medislot/pkg/utils"
)

func main() {
	
	logLevel := slog.LevelDebug
	if os.Getenv("GIN_MODE") == "release" {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := utils.NewPostgresDB(cfg.DSN())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connection established", "host", cfg.DBHost, "name", cfg.DBName)

	userRepo := repository.NewUserRepository(db)
	slotRepo := repository.NewSlotRepository(db)
	apptRepo := repository.NewAppointmentRepository(db)

	userSvc := service.NewUserService(userRepo)
	slotSvc := service.NewSlotService(slotRepo)
	apptSvc := service.NewAppointmentService(apptRepo)

	authHandler := handler.NewAuthHandler(userSvc, cfg.JWTSecret, cfg.JWTExpiryHours)
	userHandler := handler.NewUserHandler(userSvc)
	slotHandler := handler.NewSlotHandler(slotSvc)
	apptHandler := handler.NewAppointmentHandler(apptSvc)

	router := handler.SetupRouter(handler.RouterDeps{
		Auth:        authHandler,
		User:        userHandler,
		Slot:        slotHandler,
		Appointment: apptHandler,
		JWTSecret:   cfg.JWTSecret,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := worker.NewScheduler(apptRepo, cfg.WorkerIntervalSeconds, cfg.AppointmentExpiryMinutes)
	go scheduler.Start(ctx)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("MediSlot API started", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	slog.Info("shutdown signal received", "signal", sig.String())
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("server exited cleanly")
}
