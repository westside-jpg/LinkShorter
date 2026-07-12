package services

import (
	"context"
	"errors"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

/*
Файл auth.go содержит функции для работы с пользователями и аутентификацией.
Включает регистрацию пользователей, авторизацию, подтверждение электронной
почты, получение информации о пользователях, работу с JWT и различные проверки,
связанные с учетными записями
*/

/*
IsUsernameInDatabase проверяет, существует ли пользователь
с указанным именем в базе данных
*/
func IsUsernameInDatabase(db *pgxpool.Pool, username string) (bool, error) {
	var exist string

	err := db.QueryRow(
		context.Background(),
		`SELECT username FROM users WHERE username = $1`,
		username,
	).Scan(&exist)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

/*
IsEmailInDatabase проверяет, зарегистрирован ли
указанный адрес электронной почты в базе данных.
*/
func IsEmailInDatabase(db *pgxpool.Pool, email string) (bool, error) {
	var exist string

	err := db.QueryRow(
		context.Background(),
		`SELECT email FROM users WHERE email = $1`,
		email,
	).Scan(&exist)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

/*
IsLoginValid проверяет корректность логина и пароля.
Функция ищет пользователя по имени или электронной почте
и сравнивает переданный пароль с сохранённым хешем
*/
func IsLoginValid(db *pgxpool.Pool, loginInput string, password string) (bool, error) {
	var hashedPassword string

	err := db.QueryRow(
		context.Background(),
		`SELECT password FROM users WHERE username = $1 OR email = $1`,
		loginInput,
	).Scan(&hashedPassword)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil // Пользователь не найден
	} else if err != nil {
		return false, err // Ошибка БД
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, nil // Пароль неверный
	}

	return true, nil // Все хорошо
}

/*
RegisterUser создаёт нового пользователя в базе данных,
генерирует код подтверждения почты и сохраняет его вместе
с данными пользователя
*/
func RegisterUser(db *pgxpool.Pool, username string, email string, password string) error {
	var verificationCode string

	verificationCode = uuid.New().String()

	_, err := db.Exec(
		context.Background(),
		`INSERT INTO users (username, password, email, verification_code) VALUES ($1, $2, $3, $4)`,
		username, password, email, verificationCode,
	)

	if err != nil {
		return err
	}

	return nil
}

/*
GetUserID возвращает идентификатор пользователя
по имени пользователя или адресу электронной почты
*/
func GetUserID(db *pgxpool.Pool, loginInput string) (int, error) {
	var userID int

	err := db.QueryRow(
		context.Background(),
		`SELECT id FROM users WHERE username = $1 OR email = $1`,
		loginInput,
	).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

/*
VerifyEmail подтверждает адрес электронной почты пользователя
по коду подтверждения
*/
func VerifyEmail(db *pgxpool.Pool, code string) error {
	_, err := db.Exec(
		context.Background(),
		`UPDATE users SET is_verified = true WHERE verification_code = $1`,
		code,
	)

	if err != nil {
		return err
	}

	return nil
}

/*
GetUserIDByCode возвращает идентификатор пользователя
по коду подтверждения электронной почты
*/
func GetUserIDByCode(db *pgxpool.Pool, code string) (int, error) {
	var userID int

	err := db.QueryRow(
		context.Background(),
		`SELECT id FROM users WHERE verification_code = $1`,
		code).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

/*
GetUserIdFromJWT извлекает идентификатор пользователя
из JWT-токена, сохранённого в cookie
*/
func GetUserIdFromJWT(c *gin.Context) (int, error) {
	tokenString, err := c.Cookie("token")
	if err != nil {
		return 0, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	return userID, nil
}
