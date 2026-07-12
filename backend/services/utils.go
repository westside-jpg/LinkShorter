package services

import (
	"fmt"
	"math/rand"
	"unicode"
)

/*
Файл utils.go содержит вспомогательные функции общего
назначения, не привязанные к конкретной сущности проекта.
Рандомизация регистра символов для коротких ссылок,
генерация кода подтверждения и склонение русских слов
по числам
*/

/*
ToUpperAndLower случайным образом преобразует
каждую букву строки в верхний или нижний регистр
и возвращает получившуюся строку
*/
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
GenerateVerificationCode генерирует случайный
шестизначный код подтверждения и возвращает его
в виде строки
*/
func GenerateVerificationCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

/*
DeclinationWord возвращает правильную форму слова
в зависимости от переданного числа согласно
правилам склонения именительного падежа русского языка
*/
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
