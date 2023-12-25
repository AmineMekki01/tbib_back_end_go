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

	r.GET("/folders/:folderId/subfolders", func(c *gin.Context) {
		services.GetSubfolders(c, pool)
	})

	r.GET("/folders/:folderId/breadcrumbs", func(c *gin.Context) {	
		services.GetBreadcrumbs(c, pool)
	})
	
	r.DELETE("/delete-files/:folderId", func(c *gin.Context) {
    	services.DeleteFolderAndContents(c, pool)
	})

	r.PATCH("/update-folder/:folderId", func(c *gin.Context) {
		services.UpdateFolderName(c, pool)
	})

	r.POST("/upload-file", func(c *gin.Context) {
		services.UploadFile(c, pool)
	})

	r.Static("/files", "./uploads")

	r.GET("/download-file/:fileId", func(c *gin.Context) {
        services.DownloadFile(c, pool)
    })

}