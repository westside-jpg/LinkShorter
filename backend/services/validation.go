package services

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

/*
Файл validation.go содержит функции проверки корректности
пользовательского ввода. Формат email, требования к паролю
и имени пользователя, сборная проверка данных при регистрации,
а также валидация ссылок при создании коротких и кастомных URL
*/

/*
IsValidEmail проверяет корректность введенной
почты и возвращает булевое значение
*/
func IsValidEmail(email string) bool {
	return regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`).MatchString(email)
}

/*
IsPasswordStrong проверяет, что в пароле есть как минимум
одна заглавная буква, одна маленькая и цифра и возвращает булево значение
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
ValidateURL проверяет и нормализует введённую пользователем
ссылку. Убирает пробелы по краям, при отсутствии схемы
добавляет https, проверяет что схема http или https и что
хост содержит точку. Возвращает нормализованную ссылку
либо текст ошибки
*/
func ValidateURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("Cначала введите ссылку")
	}
	if !strings.HasPrefix(rawURL, "http://") &&
		!strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", errors.New("Некорректная ссылка")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("Некорректная ссылка")
	}
	if !strings.Contains(parsed.Hostname(), ".") {
		return "", errors.New("Некорректная ссылка")
	}
	return rawURL, nil
}

/*
ValidateCustomLink проверяет корректность пользовательского
названия для кастомной короткой ссылки. Убирает пробелы
по краям, проверяет что название не пустое, не длиннее
25 символов и содержит только буквы, цифры, дефис и
подчёркивание. Возвращает очищенное название либо
текст ошибки
*/
func ValidateCustomLink(custom string) (string, error) {
	custom = strings.TrimSpace(custom)
	if custom == "" {
		return "", errors.New("Введите название ссылки")
	}
	if len([]rune(custom)) > 25 {
		return "", errors.New("Название ссылки не должно превышать 25 символов")
	}
	reg := regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)
	if !reg.MatchString(custom) {
		return "", errors.New("Название ссылки может содержать только буквы, цифры, '_' и '-'")
	}
	return custom, nil
}
