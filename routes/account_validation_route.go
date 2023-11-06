package routes

import (
	"tbibi_back_end_go/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func SetupAccountValidationRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/activate_account", func(c *gin.Context) {
		services.ActivateAccount(c, pool)
		})
}

