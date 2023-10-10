package auth

import (
	"github.com/dgrijalva/jwt-go"
)

type User struct {
	ID string `json:"id"`
}


func GenerateToken(user User, userType string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   user.ID,
		"userType": userType,
	})

	tokenString, err := token.SignedString([]byte("your-secret"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}