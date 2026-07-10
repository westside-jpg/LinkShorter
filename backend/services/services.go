package services

import (
	"LinkShorter/models"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

/*
GenerateLink принимает длинную ссылку, записывает (сверяет с) в БД и возвращает короткую
для случаев когда пользователь зарегистрирован или не зарегистрирован
*/
func GenerateLink(db *pgxpool.Pool, longLink string, userID int) (string, error) {
	var link = "localhost:8080/"
	var existedLink string

	if userID == 0 {
		err := db.QueryRow(
			context.Background(),
			`SELECT short_url FROM links
				WHERE original_url = $1 AND user_id = 0`,
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

	err := db.QueryRow(
		context.Background(),
		`SELECT short_url FROM links
				WHERE original_url = $1 AND user_id = $2`,
		longLink, userID,
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
						 (original_url, short_url, user_id) VALUES ($1, $2, $3)`,
					longLink, link, userID,
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

	_, err = db.Exec(
		context.Background(),
		`UPDATE users SET last_send = $1
		WHERE email = $2`,
		time.Now().UTC(), email)

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

/*
VerifyEmail находит почту по коду из письма подтверждает ее
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
GetUserIDByCode возвращает id пользователя
по его коду верификации почты
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
DataAboutUserFromJWT возвращает значения для ручки /api/me
с информацией по пользователю
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

func CouldResendEmail(db *pgxpool.Pool, email string) (bool, time.Duration, error) {
	var lastSendTime time.Time

	err := db.QueryRow(
		context.Background(),
		`SELECT last_send FROM users WHERE email = $1`,
		email).Scan(&lastSendTime)

	if err != nil {
		return false, 0, err
	}

	elapsed := time.Since(lastSendTime)
	remaining := time.Minute - elapsed

	if time.Since(lastSendTime) < time.Minute {
		return false, remaining, nil
	}

	return true, 0, nil
}

func UserLinks(db *pgxpool.Pool, userId int, sortBy string, orderBy string) ([]models.Link, error) {
	var sortColumn string
	switch sortBy {
	case "date":
		sortColumn = "created_at"
	case "views":
		sortColumn = "views"
	default:
		return nil, errors.New("Неверная сортировка")
	}

	var sortOrder string
	switch orderBy {
	case "asc":
		sortOrder = "ASC"
	case "desc":
		sortOrder = "DESC"
	default:
		return nil, errors.New("Неверный порядок сортировки")
	}

	query := fmt.Sprintf(`
		SELECT id,
		       short_url,
		       original_url,
		       views,
		       tag,
		       created_at
		FROM links
		WHERE user_id = $1
		ORDER BY %s %s
	`, sortColumn, sortOrder)

	rows, err := db.Query(
		context.Background(),
		query,
		userId,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var links []models.Link
	for rows.Next() {
		var link models.Link
		if err := rows.Scan(&link.LinkID, &link.ShortURL, &link.OriginalURL, &link.Views, &link.Tag, &link.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, rows.Err()
}

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

func ScheduleDeletingNotVerifiedUsers(db *pgxpool.Pool) ([]string, error) {
	rows, err := db.Query(context.Background(),
		`SELECT email FROM users
         WHERE is_verified = FALSE
         AND created_at < NOW() - INTERVAL '3 days'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	if len(emails) == 0 {
		return nil, nil
	}

	_, err = db.Exec(context.Background(),
		`DELETE FROM users
         WHERE is_verified = FALSE
         AND created_at < NOW() - INTERVAL '3 days'`)
	if err != nil {
		return nil, err
	}

	return emails, nil
}

func SendAccountDeletingEmail(email string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Аккаунт LinkShorter был удалён")

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
									Аккаунт удалён
								</h1>
	
								<p style="margin:24px 0 16px;color:#555;font-size:16px;line-height:1.6;">
									К сожалению, Вы не подтвердили адрес электронной почты
									в течение <strong>3 дней</strong>, поэтому Ваш аккаунт был автоматически удалён...
								</p>

								<div style="display:inline-block;background:#f3f4f6;padding:16px 24px;border-radius:8px;color:#444;font-size:15px;line-height:1.6;">
									Нам очень жаль, что так получилось
								</div>
	
								<p style="margin:24px 0 0;color:#555;font-size:16px;line-height:1.6;">
									Если Вы всё ещё хотите пользоваться <strong>LinkShorter</strong>,
									просто зарегистрируйтесь снова и подтвердите адрес электронной почты
								</p>
	
								<hr style="margin:32px 0;border:none;border-top:1px solid #e5e7eb;">
	
								<p style="margin:0;color:#999;font-size:13px;line-height:1.5;">
									Спасибо, что проявили интерес к LinkShorter. Будем рады видеть Вас снова!
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

func IncreaseLinkViews(db *pgxpool.Pool, code string) error {
	_, err := db.Exec(
		context.Background(),
		`UPDATE links
		SET views = views + 1
		WHERE short_url LIKE '%' || $1 || '%'`,
		code)

	if err != nil {
		return err
	}

	return nil
}

func GenerateVerificationCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func PasswordResetEmail(email string, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Код подтверждения LinkShorter")
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
									Для подтверждения адреса электронной почты,
									чтобы сменить пароль от аккаунта, используйте следующий код
								</p>
								<div style="
									display:inline-block;
									background:#2563eb;
									color:#ffffff;
									font-size:34px;
									font-weight:bold;
									letter-spacing:8px;
									padding:18px 36px;
									border-radius:10px;
									margin:16px 0;">
									`+code+`
								</div>
								<p style="margin:24px 0 0;color:#777;font-size:14px;">
									Код действителен <strong>15 минут</strong>
								</p>
								<hr style="margin:32px 0;border:none;border-top:1px solid #e5e7eb;">
								<p style="margin:0;color:#999;font-size:13px;line-height:1.5;">
									Если Вы не запрашивали этот код, просто проигнорируйте это письмо
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

func AddPasswordResetCodeToDatabase(db *pgxpool.Pool, email string, code string) error {
	_, err := db.Exec(
		context.Background(),
		`UPDATE users SET reset_password_code=$1, last_send=$2, reset_password_attempts=5 WHERE email=$3`,
		code, time.Now().UTC(), email)

	if err != nil {
		return err
	}

	return nil
}

func CheckResetPasswordCode(db *pgxpool.Pool, email string, code string) (bool, error) {
	var userID int
	err := db.QueryRow(
		context.Background(),
		`SELECT id FROM users WHERE email=$1 AND reset_password_code=$2`,
		email, code).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil

}

func IsCodeFresh(db *pgxpool.Pool, email string) (bool, error) {
	var valid bool
	err := db.QueryRow(
		context.Background(),
		`SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE email = $1
			  AND NOW() < last_send + INTERVAL '15 minutes'
		)`, email).Scan(&valid)

	if err != nil {
		return false, err
	}

	return valid, nil
}

func DecreaseResetPasswordAttempts(db *pgxpool.Pool, email string) (int, error) {
	var attempts int

	err := db.QueryRow(
		context.Background(),
		`UPDATE users
		SET reset_password_attempts = GREATEST(reset_password_attempts - 1, 0)
		WHERE email = $1
		RETURNING reset_password_attempts`,
		email,
	).Scan(&attempts)

	if err != nil {
		return 0, err
	}

	return attempts, nil
}

func DeclinationWord(n int, one, two, many string) string {
	lastTwoDigits := n % 100

	if lastTwoDigits >= 11 && lastTwoDigits <= 14 {
		return many
	}

	switch n % 10 {
	case 1:
		return one
	case 2, 3, 4:
		return two
	default:
		return many
	}
}

func ResetPassword(db *pgxpool.Pool, email string, password string) error {
	_, err := db.Exec(
		context.Background(),
		`UPDATE users SET password=$1 WHERE email=$2`,
		password, email)

	if err != nil {
		return err
	}

	return nil
}

func DeleteLink(db *pgxpool.Pool, linkID int, userID int) error {
	_, err := db.Exec(
		context.Background(),
		`DELETE FROM links WHERE id=$1 AND user_id=$2`,
		linkID, userID)

	if err != nil {
		return err
	}

	return nil
}

func AddTag(db *pgxpool.Pool, linkID int, tag string) error {
	_, err := db.Exec(
		context.Background(),
		`UPDATE links SET tag=$1 WHERE id=$2`,
		tag, linkID)

	if err != nil {
		return err
	}

	return nil
}

func CouldAddTag(db *pgxpool.Pool, userID int, linkID int) (bool, error) {
	var valid bool
	err := db.QueryRow(
		context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM links WHERE id=$1 AND user_id=$2
			)`,
		linkID, userID).Scan(&valid)

	if err != nil {
		return false, err
	}

	return valid, nil
}
