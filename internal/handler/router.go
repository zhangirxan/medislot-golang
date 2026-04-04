package handler

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"medislot/internal/models"
	"medislot/pkg/middleware"
)

type RouterDeps struct {
	Auth        *AuthHandler
	User        *UserHandler
	Slot        *SlotHandler
	Appointment *AppointmentHandler
	JWTSecret   string
}

func SetupRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(requestLogger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/register", deps.Auth.Register)
	r.POST("/login", deps.Auth.Login)

	auth := r.Group("/", middleware.AuthMiddleware(deps.JWTSecret))

	users := auth.Group("/users")
	{
		users.GET("", middleware.RequireRole(models.RoleAdmin), deps.User.GetAll)
		users.GET("/:id", deps.User.GetByID)
		users.PUT("/:id", deps.User.Update)
		users.DELETE("/:id", middleware.RequireRole(models.RoleAdmin), deps.User.Delete)
	}

	slots := auth.Group("/slots")
	{
		slots.POST("", middleware.RequireRole(models.RoleDoctor), deps.Slot.Create)
		slots.GET("", deps.Slot.GetAvailable)
		slots.GET("/my", middleware.RequireRole(models.RoleDoctor), deps.Slot.GetMySlots)
		slots.DELETE("/:id", middleware.RequireRole(models.RoleDoctor), deps.Slot.Cancel)
	}

	appts := auth.Group("/appointments")
	{
		appts.POST("", middleware.RequireRole(models.RolePatient), deps.Appointment.Book)
		appts.GET("", deps.Appointment.GetMyAppointments)
		appts.GET("/:id", deps.Appointment.GetByID)
		appts.DELETE("/:id", deps.Appointment.Cancel)
	}

	return r
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		slog.Info("request",
			"method",  c.Request.Method,
			"path",    path,
			"status",  c.Writer.Status(),
			"latency", time.Since(start).String(),
			"ip",      c.ClientIP(),
		)
	}
}
