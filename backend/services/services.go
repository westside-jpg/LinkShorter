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

// GenerateLink принимает длиннную ссылку, записывает (сверяет с) в БД и возвращает короткую
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
