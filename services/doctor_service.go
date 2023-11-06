package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"tbibi_back_end_go/auth"
	"tbibi_back_end_go/models"
	"tbibi_back_end_go/validators"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

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
	var doctor models.Doctor

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	verificationLink := validators.GenerateVerificationLink(doctor.Email, c, pool)
	if verificationLink == "" {
		// Handle the error if the link couldn't be generated
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate verification link"})
		return
	}

	// Send the verification email
	err = validators.SendVerificationEmail(doctor.Email, verificationLink) 
	if err != nil {
		// Log the error and send a response to the user
		log.Printf("Failed to send verification email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	// Respond to the user
	c.JSON(http.StatusCreated, gin.H{
		"success": "true",
		"message": "Doctor created successfully. Please check your email to verify your account.",
	})
}

func doctorToAuthUser(d *models.Doctor) auth.User {
	return auth.User{ID: d.Email}  
}

func LoginDoctor(c *gin.Context, pool *pgxpool.Pool) {
	var loginReq models.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// check if the account is verified
	var isVerified bool
	ctx := context.Background()
	err := pool.QueryRow(ctx, "SELECT is_verified FROM doctor_info WHERE email = $1", loginReq.Email).Scan(&isVerified)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Account Not Verified"})
		return	
	}

	if !isVerified {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Account Not Verified, Please check your email to verify your account."})
		return
	}

	// Fetch the doctor from the database based on the email
	var doctor models.Doctor
	ctx = context.Background()
	err = pool.QueryRow(ctx, "SELECT email, hashed_password FROM doctor_info WHERE email = $1", loginReq.Email).Scan(
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
	err = pool.QueryRow(ctx, "SELECT doctor_id FROM doctor_info WHERE email = $1", loginReq.Email).Scan(
		&doctor.DoctorID,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "token": token, "doctor_id": doctor.DoctorID})
}



func GetDoctorById(c *gin.Context, pool *pgxpool.Pool) {
    doctorId := c.Param("doctorId")  
	var doctor models.Doctor
	doctor.DoctorID = doctorId

    err := pool.QueryRow(context.Background(), "SELECT email, phone_number, first_name, last_name, TO_CHAR(birth_date, 'YYYY-MM-DD'), doctor_bio, sex, location, specialty, rating_score, rating_count  FROM doctor_info WHERE doctor_id = $1", doctor.DoctorID).Scan(
        &doctor.Email,
        &doctor.PhoneNumber,
        &doctor.FirstName, 
        &doctor.LastName,
        &doctor.BirthDate,
        &doctor.DoctorBio,
        &doctor.Sex,
		&doctor.Location,
		&doctor.Specialty,
		&doctor.RatingScore,
		&doctor.RatingCount,
    )
    
    if err != nil {
        if err.Error() == "no rows in result set" {
            c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
        } else {
            log.Println("Database error:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
        }
        return
    }

    c.JSON(http.StatusOK, doctor) 
}



func GetAllDoctors(c *gin.Context, pool *pgxpool.Pool) {
	var doctors []models.Doctor
	query := c.DefaultQuery("query", "")
	specialty := c.DefaultQuery("specialty", "")
	location := c.DefaultQuery("location", "")

	sqlQuery := "SELECT doctor_id, username, first_name, last_name, specialty, experience, rating_score,rating_count, location FROM doctor_info"
	var conditions []string
	var queryParams []interface{}

	if query != "" || specialty != "" || location != "" {
		sqlQuery += " WHERE "
		if query != "" {
			conditions = append(conditions, fmt.Sprintf("(first_name ILIKE $%d OR last_name ILIKE $%d)", len(queryParams)+1, len(queryParams)+1))
			queryParams = append(queryParams, "%"+query+"%", "%"+query+"%")
		}
		if specialty != "" {
			conditions = append(conditions, fmt.Sprintf("specialty ILIKE $%d", len(queryParams)+1))
			queryParams = append(queryParams, "%"+specialty+"%")
		}
		if location != "" {
			conditions = append(conditions, fmt.Sprintf("location ILIKE $%d", len(queryParams)+1))
			queryParams = append(queryParams, "%"+location+"%")
		}
		sqlQuery += strings.Join(conditions, " AND ")
	}

	rows, err := pool.Query(context.Background(), sqlQuery, queryParams...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var doctor models.Doctor
		err := rows.Scan(
			&doctor.DoctorID, 
			&doctor.Username, 
			&doctor.FirstName, 
			&doctor.LastName, 
			&doctor.Specialty, 
			&doctor.Experience, 
			&doctor.RatingScore, 
			&doctor.RatingCount,  
			&doctor.Location,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}
		log.Println(doctor)
		doctors = append(doctors, doctor)
	}

	c.JSON(http.StatusOK, doctors)
}
