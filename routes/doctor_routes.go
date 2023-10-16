package routes

import (
	"tbibi_back_end_go/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)



func SetupDoctorRoutes(r *gin.Engine, pool *pgxpool.Pool) {

	r.POST("/api/v1/doctors/register", func(c *gin.Context) {
		services.RegisterDoctor(c, pool)
	})

	r.POST("/api/v1/doctors/login", func(c *gin.Context) {
		services.LoginDoctor(c, pool)
	})

	r.GET("/api/v1/doctors/:doctorId", func(c *gin.Context) {
		services.GetDoctorById(c, pool)
	})

	r.GET("/api/v1/doctors", func(c *gin.Context) {
		services.GetAllDoctors(c, pool)
	})
	
}
