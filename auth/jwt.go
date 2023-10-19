package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type User struct {
	ID string `json:"id"`
}


func GenerateToken(user User, userType string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   user.ID,
		"userType": userType,
		"exp":      expirationTime.Unix(),
	  })

	tokenString, err := token.SignedString([]byte("your-secret"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}