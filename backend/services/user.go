package services

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Файл user.go содержит функции для получения данных
о пользователе из базы данных
*/

/*
DataAboutUserFromJWT возвращает основные данные пользователя
из базы данных по его id (username, email, статус верификации
почты, дата регистрации). Используется для ручки /api/me,
куда userId передаётся уже извлечённым из JWT
*/
func DataAboutUserFromJWT(db *pgxpool.Pool, userId int) (string, string, bool, time.Time, error) {
	var username string
	var email string
	var isVerified bool
	var createdAt time.Time

	err := db.QueryRow(
		context.Background(),
		`SELECT username, email, is_verified, created_at FROM users WHERE id = $1`,
		userId).Scan(&username, &email, &isVerified, &createdAt)

	if err != nil {
		return "", "", false, time.Time{}, err
	}

	return username, email, isVerified, createdAt, nil
}
