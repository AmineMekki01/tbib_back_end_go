package main

import (
	"fmt"
	"log"
	"tbibi_back_end_go/db"
	"tbibi_back_end_go/routes"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(config))

	// Initialize database
	conn, err := db.InitDatabase()

	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	
	defer conn.Close()

	// Initialize routes
	routes.SetupPatientRoutes(r, conn)
	routes.SetupDoctorRoutes(r, conn)

	r.Use(func(c *gin.Context) {
        for k, v := range c.Writer.Header() {
            fmt.Println(k, v)
        }
    })
    // Start server
    r.Run(":3001")
}