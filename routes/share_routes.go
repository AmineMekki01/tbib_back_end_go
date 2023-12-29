package routes

import (
	"tbibi_back_end_go/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func SetupShareRoutes(r *gin.Engine, pool *pgxpool.Pool) {

	r.POST("/api/v1/share", func(c *gin.Context) {
		services.ShareItem(c, pool)
	})

	r.GET("/api/v1/shared-with-me", func(c *gin.Context) {
		services.GetSharedWithMe(c, pool)
	})

	r.GET("/api/v1/shared-by-me", func(c *gin.Context) {
		services.GetSharedByMe(c, pool)
	})

	r.GET("/api/v1/doctors-to-share-with", func(c *gin.Context) {
        services.ListDoctors(c, pool)
    })
}	