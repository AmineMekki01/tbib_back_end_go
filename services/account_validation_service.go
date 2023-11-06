package services

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)


func ActivateAccount(c *gin.Context, pool *pgxpool.Pool) {
    token := c.Query("token")
    log.Println("Got the token:", token)
    if token == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
        return

    }

    var email string
    err := pool.QueryRow(context.Background(), "SELECT email FROM verification_tokens WHERE token = $1", token).Scan(&email)
    if err != nil {		
		log.Println("Token is no longer valid")
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Token is no longer valid"})
    }
    
    log.Println("Got the email:", email)
    
    var patientId string
    err = pool.QueryRow(context.Background(), "SELECT patient_id FROM patient_info WHERE email = $1", email).Scan(&patientId)
    if err != nil {
        log.Println("Error retrieving patient ID:", err)
        if err.Error() == "no rows in result set" {
            log.Println("User is not a patient.")
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }
    } else {
        log.Println("Got the patient id:", patientId)
    }

    var doctorId string
    err = pool.QueryRow(context.Background(), "SELECT doctor_id FROM doctor_info WHERE email = $1", email).Scan(&doctorId)
    if err != nil {
        log.Println("Error retrieving doctor ID:", err)
        if err.Error() == "no rows in result set" {
            log.Println("User is not a doctor.")
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }
    } else {
        log.Println("Got the doctor id:", doctorId)
    }

	
	log.Println("Got the patient id")
	log.Println(patientId)
	log.Println("Got the doctor id")
	log.Println(doctorId)

	if patientId != "" {
		_, err = pool.Exec(context.Background(), "UPDATE patient_info SET is_verified = true WHERE email = $1", email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	} else if doctorId != "" {
		_, err = pool.Exec(context.Background(), "UPDATE doctor_info SET is_verified = true WHERE email = $1", email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}


	// Optionally, delete the token so it can't be used again
    _, err = pool.Exec(context.Background(), "DELETE FROM verification_tokens WHERE token = $1", token)
    if err != nil {
        // Handle error, but it's not fatal so don't return
        log.Printf("Failed to delete verification token: %v", err)
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "message": "Account activated successfully"})
}
