package services

import (
	"context"
	"errors"
	"math/rand"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ToUpperAndLower делает буквы слова то маленькими, то большими
func ToUpperAndLower(text string) string {
	runes := []rune(text)

	for i := len(runes) - 1; i >= 0; i-- {
		var randValue = rand.Intn(2)

		if randValue == 0 {
			runes[i] = unicode.ToUpper(runes[i])
		} else {
			runes[i] = unicode.ToLower(runes[i])
		}

	}

	return string(runes)
}

// GenerateLink принимает длинную ссылку, записывает (сверяет с) в БД и возвращает короткую
func GenerateLink(db *pgxpool.Pool, longLink string) (string, error) {
	var link = "localhost:8080/"

	var existedLink string

	err := db.QueryRow(
		context.Background(),
		`SELECT short_url FROM links
			WHERE original_url = $1`,
		longLink,
	).Scan(&existedLink)

	if errors.Is(err, pgx.ErrNoRows) {
		var id int
		for {
			var randomEnd = ToUpperAndLower((uuid.New()).String())[0:6]

			err := db.QueryRow(
				context.Background(),
				`SELECT id FROM links
			WHERE short_url = $1`,
				link+randomEnd,
			).Scan(&id)

			if errors.Is(err, pgx.ErrNoRows) {
				link = link + randomEnd
				_, err = db.Exec(
					context.Background(),
					`INSERT INTO links
				(original_url, short_url) VALUES ($1, $2)`,
					longLink, link,
				)

				if err != nil {
					return "", err
				}

				return link, nil

			} else if err != nil {
				return "", err
			}

		}
	} else if err != nil {
		return "", err
	}

	return existedLink, nil
}

/*
ReturnOriginalLink принимает на вход базу данных и код из
короткой ссылки, возвращает оригинальную ссылку из базы данных
*/
func ReturnOriginalLink(db *pgxpool.Pool, code string) (string, error) {
	var originalURL string
	var shortURL = "localhost:8080/" + code

	err := db.QueryRow(
		context.Background(),
		`SELECT original_url FROM links WHERE short_url = $1`,
		shortURL,
	).Scan(&originalURL)

	if err != nil {
		return "", err
	}

	return originalURL, nil
}

/*
IsUsernameInDatabase проверяет существование имени пользователя
в базе данных и возращает булевое значение
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
IsEmailInDatabase проверяет существование почты
в базе данных и возращает булевое значение
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
IsLoginValid проверяет существование пользователя
в базе данных и возращает булевое значение
*/
func IsLoginValid(db *pgxpool.Pool, username string, password string) (bool, error) {
	var exist string

	err := db.QueryRow(
		context.Background(),
		`SELECT 1 FROM users WHERE username = $1, password = $2`,
		username, password,
	).Scan(&exist)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil

	// Дописать проверку почта + пароль
}

/*
RegisterUser добавляет в базу данных username, email, password, verification_code
зарегистрированного пользователя и возвращает ошибку, если она есть
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
