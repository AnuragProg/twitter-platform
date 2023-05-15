package utils

import (
	"errors"
	"os"
	"strings"
	"time"
	"github.com/golang-jwt/jwt/v5"
)


func GenerateToken(username, userId string) (string, error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"isAdmin": strings.ToLower(username) == "anurag",
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

/* @returns userId, isAdmin, error*/
func VerifyToken(tokenString string)(string, bool, error){
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok{
			return nil, errors.New("invalid token")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err!=nil{
		return "", false, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); token.Valid && ok{
		return claims["userId"].(string), claims["isAdmin"].(bool), nil
	}
	return "", false, errors.New("invalid token")
}