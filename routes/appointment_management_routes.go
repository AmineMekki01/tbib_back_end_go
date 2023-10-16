package routes

import (
	"tbibi_back_end_go/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)




func SetupAppointmentManagementRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/api/v1/availabilities", func(c *gin.Context) {
		services.GetAvailabilities(c, pool)
	})

	r.POST("/api/v1/reservations", func(c *gin.Context) {
		services.CreateReservation(c, pool)
	})


	r.GET("/api/v1/reservations", func(c *gin.Context) {
		services.GetReservations(c, pool)
	})

}

