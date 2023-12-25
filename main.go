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
		AllowOrigins: []string{"http://localhost:3000"},
        AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders: []string{"Origin", "Content-Type", "Content-Length"},
        ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(config))

    // Middleware for debugging headers
    r.Use(func(c *gin.Context) {
        // Print request headers
        fmt.Printf("Request %v Headers:\n", c.Request.Method)
        for k, v := range c.Request.Header {
            fmt.Println(k, v)
        }
        c.Next()
        
        // Print response headers
        fmt.Printf("Response Headers:\n")
        for k, v := range c.Writer.Header() {
            fmt.Println(k, v)
        }
    })

	// Initialize database
	conn, err := db.InitDatabase()

	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	
	defer conn.Close()

	// Initialize routes
	routes.SetupPatientRoutes(r, conn)
	routes.SetupDoctorRoutes(r, conn)
	routes.SetupAppointmentManagementRoutes(r, conn)
	routes.SetupFileRoutes(r, conn)
	routes.SetupAccountValidationRoutes(r, conn)


	r.Use(func(c *gin.Context) {
        for k, v := range c.Writer.Header() {
            fmt.Println(k, v)
        }
    })
    // Start server
    r.Run(":3001")
}
