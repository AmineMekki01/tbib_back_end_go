package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)


func ActivateAccount(c *gin.Context, pool *pgxpool.Pool) {
    token := c.Query("token")
    if token == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
        return

    }

    var email string
    err := pool.QueryRow(context.Background(), "SELECT email FROM verification_tokens WHERE token = $1 AND type = $2", token, "Account Validation").Scan(&email)
    if err != nil {		
		log.Println("Token is no longer valid")
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Token is no longer valid"})
    }
        
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


    _, err = pool.Exec(context.Background(), "DELETE FROM verification_tokens WHERE token = $1", token)
    if err != nil {
        log.Printf("Failed to delete verification token: %v", err)
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "message": "Account activated successfully"})
}


const TokenLength = 64 

//  Gnerate a secure random hex string
func GenerateSecureToken() (string, error) {
    bytes := make([]byte, TokenLength/2)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}


// Send reset password email
func SendResetPasswordEmail(recipientEmail, verificationLink string) error {

	err := godotenv.Load("./.env")
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_EMAIL_PASSWORD")
	to := []string{recipientEmail}
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	subject := "Reset your TBIBI app password."
	body := "Please click on the on the link below to reset your password:\n" + verificationLink

	message := []byte("From: " + from + "\n" +
		"To: " + recipientEmail + "\n" +
		"Subject: " + subject + "\n\n" +
		body)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return fmt.Errorf("failed to send reset password email: %v", err)
	}
	return nil
}



// RequestReset handles the initiation of the password reset process
func RequestReset(c *gin.Context, pool *pgxpool.Pool) {
    var requestBody struct{
        Email string `json:"email"`
        UserType string `json:"localUserType"`
    }
    if err := c.BindJSON(&requestBody); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    var table string
    if requestBody.UserType == "doctor" {
        table = "doctor_info"
    } else {
        table = "patient_info"
    }

    var email string
    query := fmt.Sprintf("SELECT email FROM %s WHERE email = $1", table)
    err := pool.QueryRow(c, query, requestBody.Email).Scan(&email)

    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
        return
    }

    token, err := GenerateSecureToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate a secure token"})
        return
    }

    _, err = pool.Exec(c, "INSERT INTO verification_tokens (token, email, type) VALUES ($1, $2, $3)", token, email, "Password Reset")

    resetLink := "https://localhost:3000/reset-password?token=" + token

    err = SendResetPasswordEmail(email, resetLink)  
    if err != nil { 
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset password email"})   
        return
    }   

    c.JSON(http.StatusOK, gin.H{"message": "If your email is in our system, you will receive a password reset link shortly."})
}


func UpdatePassword(c *gin.Context, pool *pgxpool.Pool) {
    var requestBody struct {
        Token       string
        NewPassword string
    }
    if err := c.BindJSON(&requestBody); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    query := fmt.Sprintf("SELECT email FROM verification_tokens WHERE token = $1 AND type = $2")
    var email string
    err := pool.QueryRow(c, query, requestBody.Token, "Password Reset").Scan(&email)
    if err != nil { 
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})    
        return
    }   

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    var table string
    var id string   
    query = fmt.Sprintf("SELECT patient_id FROM patient_info WHERE email = $1")
    err = pool.QueryRow(c, query, email).Scan(&id)  
    if err != nil {     
        if err.Error() == "no rows in result set" { 
            query = fmt.Sprintf("SELECT doctor_id FROM doctor_info WHERE email = $1")   
            err = pool.QueryRow(c, query, email).Scan(&id)  
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})

                return  
            }       
            table = "doctor_info"
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }
    } else {
        table = "patient_info"
    }

    query = fmt.Sprintf("UPDATE %s SET hashed_password = $1 WHERE email = $2", table)
    _, err = pool.Exec(c, query, hashedPassword, email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"}) 
        return
    }

    _, err = pool.Exec(c, "DELETE FROM verification_tokens WHERE token = $1", requestBody.Token)
    if err != nil {
        log.Printf("Failed to delete verification token: %v", err)
    }


    c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password has been reset successfully."})
}