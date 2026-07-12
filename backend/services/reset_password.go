package services

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Файл reset_password.go содержит функции для полного цикла
сброса пароля. Отправка и сохранение кода подтверждения,
проверка кода и его срока действия, учёт оставшихся попыток
ввода и итоговое обновление пароля в базе данных
*/

/*
AddPasswordResetCodeToDatabase сохраняет код
для сброса пароля, время его отправки и
сбрасывает количество оставшихся попыток
*/
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

/*
CheckResetPasswordCode проверяет соответствие
кода сброса пароля указанному адресу
электронной почты
*/
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

/*
IsCodeFresh проверяет, не истёк ли срок
действия кода сброса пароля
*/
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

/*
DecreaseResetPasswordAttempts уменьшает
количество оставшихся попыток ввода
кода для сброса пароля и возвращает их остаток
*/
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

/*
ResetPassword обновляет пароль пользователя
для указанного адреса электронной почты
*/
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
