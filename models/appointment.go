package models

import "time"

type Availability struct {
	AvailabilityID    int       `json:"AvailabilityId"`
	AvailabilityStart time.Time `json:"AvailabilityStart"`
	AvailabilityEnd   time.Time `json:"AvailabilityEnd"`
	DoctorID          string    `json:"DoctorId"`
}

type Reservation struct {
	ReservationID    int       `json:"reservation_id"`
	ReservationStart time.Time `json:"reservation_start"`
	ReservationEnd   time.Time `json:"reservation_end"`
	DoctorFirstName  string    `json:"doctor_first_name"`
	DoctorLastName   string    `json:"doctor_last_name"`
	Specialty        string    `json:"specialty"`
	PatientFirstName string    `json:"patient_first_name"`
	PatientLastName  string    `json:"patient_last_name"`
	Age              int       `json:"age"`
	PatientID        string    `json:"patient_id"`
	DoctorID         string    `json:"doctor_id"`
}
