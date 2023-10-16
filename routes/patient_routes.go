package routes

import (
	"tbibi_back_end_go/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func SetupPatientRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/api/v1/patients/:patientId", func(c *gin.Context) {
		services.GetPatientById(c, pool)
	})

	r.POST("/api/v1/patients/register", func(c *gin.Context) {
		services.RegisterPatient(c, pool)  
	})

	r.POST("/api/v1/patients/login", func(c *gin.Context) {
		services.LoginPatient(c, pool)
	})
}
