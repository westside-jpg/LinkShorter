package services

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
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

// SendEmail отправляет ссылку для подтверждения почты
func SendEmail(db *pgxpool.Pool, email string) error {
	var verificationCode string

	err := db.QueryRow(
		context.Background(),
		`SELECT verification_code FROM users WHERE email = $1`,
		email,
	).Scan(&verificationCode)

	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Подтверждение почты для LinkShorter")
	m.SetBody("text/html", `
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
	</head>
	<body style="margin:0;padding:0;background:#f4f4f7;font-family:Arial,Helvetica,sans-serif;">
		<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f7;padding:40px 0;">
			<tr>
				<td align="center">
					<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;padding:40px;box-shadow:0 4px 16px rgba(0,0,0,.08);">
						<tr>
							<td align="center">
								<h1 style="margin:0;color:#333;font-size:28px;">
									Подтверждение почты
								</h1>

								<p style="margin:24px 0 16px;color:#555;font-size:16px;line-height:1.6;">
									Спасибо за регистрацию! Для подтверждения адреса электронной почты
									нажмите на кнопку ниже
								</p>

                            	<a href="http://localhost:8080/verify/`+verificationCode+`"
                               		style="display:inline-block;background:#2563eb;color:#ffffff;
                                    	  text-decoration:none;padding:14px 32px;
                                    	  border-radius:8px;font-size:16px;font-weight:bold;">
                                	Подтвердить почту
                            	</a>

								<p style="margin:24px 0 0;color:#777;font-size:14px;">
									Ссылка действительна <strong>3 дня</strong>
								</p>

                            	<hr style="margin:32px 0;border:none;border-top:1px solid #e5e7eb;">

								<p style="margin:0;color:#999;font-size:13px;line-height:1.5;">
									Если вы не регистрировались на нашем сайте, просто проигнорируйте это письмо
								</p>
                        	</td>
                    	</tr>
                	</table>
            	</td>
       	 </tr>
    	</table>
	</body>
	</html>
	`)

	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
	)

	return d.DialAndSend(m)
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

/*
GetUserID возвращает id пользователя из БД
по его имени пользователя или почте
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
GetUserVerified возвращает булевое значение из БД
относительно верификации почты по его имени пользователя или почте
*/
func GetUserVerified(db *pgxpool.Pool, loginInput string) (bool, error) {
	var isVerified bool

	err := db.QueryRow(
		context.Background(),
		`SELECT is_verified FROM users WHERE username = $1 OR email = $1`,
		loginInput,
	).Scan(&isVerified)

	if err != nil {
		return false, err
	}

	return isVerified, nil
}

/*
GetUserEmail возвращает email пользователя из БД
по его имени пользователя или почте
*/
func GetUserEmail(db *pgxpool.Pool, loginInput string) (string, error) {
	var email string

	err := db.QueryRow(
		context.Background(),
		`SELECT email FROM users WHERE username = $1 OR email = $1`,
		loginInput,
	).Scan(&email)

	if err != nil {
		return "", err
	}

	return email, nil
}

/*
IsValidEmail проверяет корректность введенной
почты и возвращает булевое значение
*/
func IsValidEmail(email string) bool {
	return regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`).MatchString(email)
}

/*
IsPasswordStrong проверяет, что в пароле есть как минимум
одна заглавная буква, одна маленькая и цифра и возращает булево значение
*/
func IsPasswordStrong(password string) bool {
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasUpper && hasLower && hasDigit
}

/*
IsValidPassword проверяет что пароль состоит только из
латинских букв, цифр и спецсимволов (без пробелов и кириллицы)
*/
func IsValidPassword(password string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]+$`).MatchString(password)
}

/*
IsValidUsername проверяет что имя пользователя состоит только
из латинских букв, цифр и подчёркивания
*/
func IsValidUsername(username string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username)
}

/*
ValidateRegistration проверяет username, email и password на соответствие
требованиям (длина, допустимые символы, сложность пароля, корректность почты)
и возвращает список текстовых ошибок. Пустой слайс означает что данные валидны
*/
func ValidateRegistration(username string, email string, password string) []string {
	var errs []string

	if len(username) > 30 {
		errs = append(errs, "Максимальная длина имени 30 символов")
	}
	if len(username) < 3 {
		errs = append(errs, "Минимальная длина имени 3 символа")
	}
	if !IsValidUsername(username) {
		errs = append(errs, "Имя пользователя может содержать только латинские буквы, цифры и подчёркивание")
	}

	if !IsValidEmail(email) {
		errs = append(errs, "Почта введена некорректно")
	}

	if len(password) < 8 {
		errs = append(errs, "Минимальная длина пароля 8 символов")
	}
	if len(password) > 72 {
		errs = append(errs, "Максимальная длина пароля 72 символа")
	}
	if !IsValidPassword(password) {
		errs = append(errs, "Пароль содержит недопустимые символы")
	}
	if !IsPasswordStrong(password) {
		errs = append(errs, "Пароль должен содержать заглавную букву, строчную букву и цифру")
	}

	return errs
}
