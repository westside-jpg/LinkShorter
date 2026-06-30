package jwt_service

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/*
GenerateJWT создает JWT-токен с полями:
- user_id (айди пользователя в БД)
- is_verified (подтверждена ли почта)
*/
func GenerateJWT(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
