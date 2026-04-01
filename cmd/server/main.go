package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"medislot-backend/internal/config"
	"medislot-backend/internal/handler"
	"medislot-backend/internal/middleware"
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

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	api := r.Group("/api")
	{
		api.GET("/ping", healthHandler.Ping)

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/me", authHandler.Me)
		}
	}

	log.Printf("server is running on port %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}