package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"medislot-backend/internal/config"
	"medislot-backend/internal/handler"
	"medislot-backend/internal/repository"
	"medislot-backend/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	r := gin.Default()

	healthService := service.NewHealthService()
	healthHandler := handler.NewHealthHandler(healthService)

	api := r.Group("/api")
	{
		api.GET("/ping", healthHandler.Ping)
	}

	log.Printf("server is running on port %s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}
