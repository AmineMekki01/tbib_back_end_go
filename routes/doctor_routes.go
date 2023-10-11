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


type Doctor struct {
    Username      string `json:"Username"`
    FirstName     string `json:"FirstName"`
    LastName      string `json:"LastName"`
	Password      string `json:"Password"`
    Age           int `json:"age"`
    Sex           string `json:"Sex"`
    Specialty     string `json:"Specialty"`
    Experience    string `json:"Experience"`
    MedicalLicense string `json:"MedicalLicense"`
    DoctorBio     string `json:"DoctorBio"`
    Email         string `json:"Email"`
    PhoneNumber   string `json:"PhoneNumber"`
    StreetAddress string `json:"StreetAddress"`
    CityName   string `json:"CityName"`
    StateName  string `json:"StateName"`
    ZipCode       string `json:"ZipCode"`
    CountryName string `json:"CountryName"`
    BirthDate     string `json:"BirthDate"`
    Location      string `json:"location"`
}

func SetupDoctorRoutes(r *gin.Engine, pool *pgxpool.Pool) {

	r.POST("/api/v1/doctors/register", func(c *gin.Context) {
		RegisterDoctor(c, pool)
	})

	r.POST("/api/v1/doctors/login", func(c *gin.Context) {
		LoginDoctor(c, pool)
	})


	
}

func emailExists(conn *pgxpool.Conn, email string, c *gin.Context) (bool, error) {
	var existingEmail string
	err := conn.QueryRow(c, "SELECT email FROM doctor_info WHERE email = $1", email).Scan(&existingEmail)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}



// RegisterDoctor registers a new doctor
func RegisterDoctor(c *gin.Context, pool *pgxpool.Pool) {
	var doctor Doctor

	if err := c.ShouldBindJSON(&doctor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	conn, err := pool.Acquire(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not acquire database connection"})
		return
	}
	defer conn.Release()

	// checking if the gmail already exists
	exists, err := emailExists(conn, doctor.Email, c)
	if err != nil {
		log.Printf("Error checking email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	}

	// checking if username already exists
	var username string
	err = conn.QueryRow(c, "SELECT username FROM doctor_info WHERE username = $1", doctor.Username).Scan(&username)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(doctor.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	
	birthDate, err := time.Parse("2006-01-02", doctor.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	doctor.Age = time.Now().Year() - birthDate.Year()


	// Location
	doctor.Location = fmt.Sprintf("%s, %s, %s, %s, %s", doctor.StreetAddress, doctor.ZipCode, doctor.CityName, doctor.StateName, doctor.CountryName)

	_, err = conn.Exec(c, `
	INSERT INTO doctor_info (
		doctor_id, 
		username, 
		first_name, 
		last_name, 
		age, 
		sex, 
		hashed_password, 
		salt, 
		specialty, 
		experience, 
		rating_score, 
		rating_count, 
		create_at, 
		update_at, 
		medical_license, 
		doctor_bio, 
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
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
		$14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
	)`, 

	doctor.Username,
	doctor.FirstName, 
	doctor.LastName, 
	doctor.Age, 
	doctor.Sex, 
	hashedPassword, 
	salt, 
	doctor.Specialty, 
	doctor.Experience, 
	nil, 
	0, 
	time.Now(), 
	time.Now(), 
	doctor.MedicalLicense, 
	doctor.DoctorBio, 
	doctor.Email, 
	doctor.PhoneNumber, 
	doctor.StreetAddress, 
	doctor.CityName, 
	doctor.StateName, 
	doctor.ZipCode, 
	doctor.CountryName, 
	doctor.BirthDate, 
	doctor.Location,
		
	)	

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": "true", "message": "Doctor registered successfully"})
}

func doctorToAuthUser(d *Doctor) auth.User {
	return auth.User{ID: d.Email}  
}

func LoginDoctor(c *gin.Context, pool *pgxpool.Pool) {
	var loginReq LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Fetch the patient from the database based on the email
	var doctor Doctor
	ctx := context.Background()
	err := pool.QueryRow(ctx, "SELECT email, hashed_password FROM doctor_info WHERE email = $1", loginReq.Email).Scan(
	&doctor.Email,
	&doctor.Password,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid email or password"})
		return
	}
	// Comparing the stored hashed password, with the hashed version of the password that was received
	err = bcrypt.CompareHashAndPassword([]byte(doctor.Password), []byte(loginReq.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid email or password"})
		return
	}

	// generating a session token
	user := doctorToAuthUser(&doctor)
	token, err := auth.GenerateToken(user, "doctor")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	// get user id 
	var doctorId string
	err = pool.QueryRow(ctx, "SELECT doctor_id FROM doctor_info WHERE email = $1", loginReq.Email).Scan(
		&doctorId,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "token": token, "doctor_id": doctorId})
}
