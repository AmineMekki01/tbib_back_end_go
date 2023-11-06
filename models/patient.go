package models

type Patient struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
	Email    string `json:"Email"`
	// Age           int    `json:"Age"`
	PhoneNumber   string `json:"PhoneNumber"`
	FirstName     string `json:"FirstName"`
	LastName      string `json:"LastName"`
	BirthDate     string `json:"BirthDate"`
	StreetAddress string `json:"StreetAddress"`
	CityName      string `json:"CityName"`
	StateName     string `json:"StateName"`
	ZipCode       string `json:"ZipCode"`
	CountryName   string `json:"CountryName"`
	PatientBio    string `json:"PatientBio"`
	Sex           string `json:"sex"`
	// Location      string `json:"location"`
}
