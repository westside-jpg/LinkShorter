package jwt_service

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/*
Файл jwt_service.go содержит функции для генерации JWT-токенов
проекта. Основной токен авторизации (30 дней) и отдельный токен
для процедуры сброса пароля (15 минут). Парсинг и чтение токенов
происходит в services, здесь только их создание и подпись
*/

/*
GenerateJWT создаёт основной JWT-токен авторизации
с полями user_id и exp (30 дней). Минимум данных в токене,
так как все остальное достается из БД через /api/me
*/
func GenerateJWT(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

/*
GenerateResetPasswordJWT создаёт отдельный JWT-токен
для процедуры сброса пароля с полями email и purpose
(15 минут). Подписан тем же JWT_SECRET, что и основной
токен, поэтому email внутри нельзя подделать, а purpose
не даёт использовать обычный токен авторизации вместо
токена сброса пароля
*/
func GenerateResetPasswordJWT(email string) (string, error) {
	claims := jwt.MapClaims{
		"email":   email,
		"purpose": "reset_password",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
