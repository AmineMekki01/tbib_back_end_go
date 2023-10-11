package routes

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)


type Availability struct {
	AvailabilityID int `json:"AvailabilityId"`
	AvailabilityStart time.Time `json:"AvailabilityStart"`
	AvailabilityEnd   time.Time `json:"AvailabilityEnd"`
	DoctorID string  `json:"DoctorId"`
}

type Reservation struct {
	ReservationID      int       `json:"reservation_id"`
	ReservationStart   time.Time `json:"reservation_start"`
	ReservationEnd     time.Time `json:"reservation_end"`
	DoctorFirstName    string    `json:"doctor_first_name"`
	DoctorLastName     string    `json:"doctor_last_name"`
	Specialty          string    `json:"specialty"`
	PatientFirstName   string    `json:"patient_first_name"`
	PatientLastName    string    `json:"patient_last_name"`
	Age                int       `json:"age"`
}


func SetupAppointmentManagementRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/api/v1/availabilities", func(c *gin.Context) {
		GetAvailabilities(c, pool)
	})

	r.POST("/api/v1/reservations", func(c *gin.Context) {
		CreateReservation(c, pool)
	})


	r.GET("/api/v1/reservations", func(c *gin.Context) {
		GetReservations(c, pool)
	})

}


func GetAvailabilities(c *gin.Context, pool *pgxpool.Pool) {
    doctorId := c.DefaultQuery("doctorId", "")
    day := c.DefaultQuery("day", "")
    currentTime := c.DefaultQuery("currentTime", "")
    timeZone := c.DefaultQuery("timeZone", "")

	const customDateFormat = "2006-01-02" 
	dayStart, err := time.Parse(customDateFormat, day) 
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day format"})
		return
	}

    dayEnd := dayStart.AddDate(0, 0, 1)
    location, err := time.LoadLocation(timeZone)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time zone"})
        return
    }

    localCurrentTime, err := time.ParseInLocation(time.RFC3339, currentTime, location)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid current time format"})
        return
    }

    rows, err := pool.Query(context.Background(),
        "SELECT availability_id, availability_start, availability_end, doctor_id FROM availabilities WHERE doctor_id = $1 AND availability_start >= $2 AND availability_end < $3 AND availability_start >= $4",
        doctorId, dayStart, dayEnd, localCurrentTime)
	if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var availabilities []Availability 
    berlinLocation, _ := time.LoadLocation("Europe/Berlin")
    for rows.Next() {
        var availability Availability
        err := rows.Scan(&availability.AvailabilityID, &availability.AvailabilityStart, &availability.AvailabilityEnd, &availability.DoctorID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        availability.AvailabilityStart = availability.AvailabilityStart.In(berlinLocation)
        availability.AvailabilityEnd = availability.AvailabilityEnd.In(berlinLocation)
        availabilities = append(availabilities, availability)
    }

    c.JSON(http.StatusOK, availabilities)
}

type Appointments struct {
	AppointmentStart          time.Time `json:"AppointmentStart"`
	AppointmentEnd            time.Time `json:"AppointmentEnd"`
	AppointmentTitle          string    `json:"AppointmentTitle"`
	DoctorID       string    `json:"DoctorID"`
	PatientID      string    `json:"PatientID"`
	AvailabilityID int       `json:"AvailabilityID"`
}

// Implement POST /api/v1/reservations
func CreateReservation(c *gin.Context, pool *pgxpool.Pool) {
	var appointment Appointments

	if err := c.ShouldBindJSON(&appointment); err != nil {
		log.Println("Bind Error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Println("Connection Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(context.Background())
	if err != nil {
		log.Println("Transaction Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Insert reservation
	_, err = tx.Exec(context.Background(),
    "INSERT INTO appointments (appointment_start, appointment_end, appointment_title, doctor_id, patient_id) VALUES ($1::timestamp with time zone, $2::timestamp with time zone, $3, $4, $5)",
    appointment.AppointmentStart, appointment.AppointmentEnd, appointment.AppointmentTitle, appointment.DoctorID, appointment.PatientID)

	if err != nil {
		log.Println("Insert Error:", err)
		tx.Rollback(context.Background())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Delete availability
	_, err = tx.Exec(context.Background(), "DELETE FROM availabilities WHERE availability_id = $1", appointment.AvailabilityID)
	if err != nil {
		log.Println("Delete Error:", err)
		tx.Rollback(context.Background())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	tx.Commit(context.Background())
	c.JSON(http.StatusCreated, gin.H{"message": "Appointment booked and availability removed successfully"})
}



// Implement GET /api/v1/reservations
func GetReservations(c *gin.Context, pool *pgxpool.Pool) {
	doctorID := c.DefaultQuery("doctor_id", "")
	patientID := c.DefaultQuery("patient_id", "")
	timezone := c.DefaultQuery("timezone", "UTC")

	if doctorID == "" && patientID == "" {
		log.Println("Bad Request: doctor_id or patient_id required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request: doctor_id or patient_id required"})
		return
	}

	query := `
		SELECT 
			appointments.appointment_id,
			appointments.appointment_start,
			appointments.appointment_end,
			doctor_info.first_name,
			doctor_info.last_name,
			doctor_info.specialty,
			patient_info.first_name AS patient_first_name,
			patient_info.last_name AS patient_last_name,
			patient_info.age
		FROM 
			appointments
		JOIN
			doctor_info ON appointments.doctor_id = doctor_info.doctor_id
		JOIN
			patient_info ON appointments.patient_id = patient_info.patient_id
	`
	params := []interface{}{}
	if doctorID != "" {
		// log print for the type of the doctorID
		log.Println(doctorID)
		query += " WHERE appointments.doctor_id = $1"
		params = append(params, doctorID)
	} else {
		query += " WHERE appointments.patient_id = $1"
		params = append(params, patientID)
	}

	rows, err := pool.Query(context.Background(), query, params...)
	if err != nil {
		log.Println("Query Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	// print the retrieved data from the front
	log.Println(rows)
	defer rows.Close()

	var reservations []Reservation
	for rows.Next() {
		var r Reservation
		err := rows.Scan(&r.ReservationID, &r.ReservationStart, &r.ReservationEnd,
			&r.DoctorFirstName, &r.DoctorLastName, &r.Specialty,
			&r.PatientFirstName, &r.PatientLastName, &r.Age)
		if err != nil {
			log.Println("Row Scan Error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		log.Println("r", r)
		// Convert time to the specified timezone
		location, err := time.LoadLocation(timezone)
		if err != nil {
			log.Println("Timezone Error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
	
		r.ReservationStart = r.ReservationStart.In(location)
		r.ReservationEnd = r.ReservationEnd.In(location)
		log.Println("r", r)		// Append to the reservations slice
		reservations = append(reservations, r)
		
	}
	
	log.Println(reservations)
	c.JSON(http.StatusOK, reservations)
}
