package routes

import (
	"tbibi_back_end_go/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func SetupFileRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.POST("/create-folder", func(c *gin.Context) {
		services.CreateFolder(c, pool)
	})

	r.GET("/folders", func(c *gin.Context) {
		services.GetFolders(c, pool)
	})

}