package routes

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"tbibi_back_end_go/auth"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)


type Patient struct {
    Username      string `json:"Username"`
	Password      string `json:"Password"`
	Email         string `json:"Email"`
	Age           int `json:"Age"`
	PhoneNumber   string `json:"PhoneNumber"`
    FirstName     string `json:"FirstName"`
    LastName      string `json:"LastName"`
	BirthDate     string `json:"BirthDate"`
    StreetAddress string `json:"StreetAddress"`
    CityName   string `json:"CityName"`
    StateName  string `json:"StateName"`
    ZipCode       string `json:"ZipCode"`
    CountryName string `json:"CountryName"`
	PatientBio     string `json:"PatientBio"`
	Sex           string `json:"sex"`
	Location 	string `json:"location"`

    
}

type LoginRequest struct {
	Email    string `json:"email"`	
	Password string `json:"password"`
}

func SetupPatientRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/api/v1/patients/:patientId", func(c *gin.Context) {
		GetPatientById(c, pool)
	})

	r.POST("/api/v1/patients/register", func(c *gin.Context) {
		RegisterPatient(c, pool)  
	})

	r.POST("/api/v1/patients/login", func(c *gin.Context) {
		LoginPatient(c, pool)
	})
}

func GetPatientById(c *gin.Context, pool *pgxpool.Pool) {
    patientId := c.Param("patientId")  
    var patient Patient
    err := pool.QueryRow(context.Background(), "SELECT email, phone_number, first_name, last_name, TO_CHAR(birth_date, 'YYYY-MM-DD'), patient_bio, sex, location  FROM patient_info WHERE patient_id = $1", patientId).Scan(
        &patient.Email,
        &patient.PhoneNumber,
        &patient.FirstName, 
        &patient.LastName,
        &patient.BirthDate,
        &patient.PatientBio,
        &patient.Sex,
		&patient.Location,
    )
    
    if err != nil {
        if err.Error() == "no rows in result set" {
            c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
        } else {
            log.Println("Database error:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
        }
        return
    }

    c.JSON(http.StatusOK, patient) 
}


func RegisterPatient(c *gin.Context, pool *pgxpool.Pool) {
    var patient Patient
    if err := c.ShouldBindJSON(&patient); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

	conn, err := pool.Acquire(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
		return
	}
	defer conn.Release()

	// checking if the gmail already exists
	var email string
	err = conn.QueryRow(c, "SELECT email FROM patient_info WHERE email = $1", patient.Email).Scan(&email)

	if err != nil {
		if err.Error() != "no rows in result set" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	}

	// checking if username already exists
	var username string
	err = conn.QueryRow(c, "SELECT username FROM patient_info WHERE username = $1", patient.Username).Scan(&username)
	if err != nil {
		if err.Error() != "no rows in result set" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		} 

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

		
	
	// Generating Salt and Hash Password
	saltBytes := make([]byte, 16)
	_, err = rand.Read(saltBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	salt := hex.EncodeToString(saltBytes)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(patient.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	//  Age
	birthDate, err := time.Parse("2006-01-02", patient.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	patient.Age = time.Now().Year() - birthDate.Year()

	// Location
	patient.Location = fmt.Sprintf("%s, %s, %s, %s, %s", patient.StreetAddress, patient.ZipCode, patient.CityName, patient.StateName, patient.CountryName)
	_, err = conn.Exec(c, `
    INSERT INTO patient_info (
        patient_id, 
        username, 
        first_name, 
        last_name, 
        age, 
        sex, 
        hashed_password, 
        salt, 
        create_at, 
        update_at, 
        patient_bio, 
        email, 
        phone_number, 
        street_address, 
        city_name, 
        state_name, 
        zip_code, 
        country_name, 
        birth_date, 
        location
    ) 
    VALUES (
        uuid_generate_v4(), 
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
        $11, $12, $13, $14, $15, $16, $17, $18, $19
    )`, 
    patient.Username, 
    patient.FirstName, 
    patient.LastName, 
    patient.Age, 
    patient.Sex, 
    hashedPassword, 
    salt, 
    time.Now(), 
    time.Now(), 
    patient.PatientBio, 
    patient.Email, 
    patient.PhoneNumber, 
    patient.StreetAddress, 
    patient.CityName, 
    patient.StateName, 
    patient.ZipCode, 
    patient.CountryName, 
    patient.BirthDate, 
    patient.Location,
)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": "true", "message": "Patient created successfully"})

}


func patientToAuthUser(p *Patient) auth.User {
	return auth.User{ID: p.Email}  
}

func LoginPatient(c *gin.Context, pool *pgxpool.Pool) {
	var loginReq LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Fetch the patient from the database based on the email
	var patient Patient
	ctx := context.Background()
	err := pool.QueryRow(ctx, "SELECT email, hashed_password FROM patient_info WHERE email = $1", loginReq.Email).Scan(
	&patient.Email,
	&patient.Password,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "wwInvalid email or password"})
		return
	}

	// Comparing the stored hashed password, with the hashed version of the password that was received
	err = bcrypt.CompareHashAndPassword([]byte(patient.Password), []byte(loginReq.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "ssInvalid email or password"})
		return
	}

	// generating a session token
	user := patientToAuthUser(&patient)
	token, err := auth.GenerateToken(user, "patient")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	// get user id 
	var patientId string
	err = pool.QueryRow(ctx, "SELECT patient_id FROM patient_info WHERE email = $1", loginReq.Email).Scan(
		&patientId,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "token": token, "patient_id": patientId})


}
