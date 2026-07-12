package services

import (
	"LinkShorter/models"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Файл links.go содержит функции для работы со ссылками.
Создание коротких и кастомных ссылок, получение оригинального
URL по короткому коду, получение списка ссылок пользователя
с сортировкой/фильтрацией/поиском, учёт просмотров, работу
с тэгами и удаление ссылок
*/

/*
GenerateLink возвращает короткую ссылку для указанного
оригинального адреса. Если ссылка для этого URL уже
существует (в анонимном пуле при userID = 0 или в личном
пуле пользователя), возвращает существующий короткий код.
Иначе генерирует случайный уникальный код (с проверкой
на коллизии) и сохраняет новую ссылку в базе данных
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
				WHERE original_url = $1 AND user_id = $2 AND is_custom = FALSE`,
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
ReturnOriginalLink возвращает оригинальную ссылку
по коду короткой ссылки
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
UserLinks возвращает список ссылок пользователя
с учётом поиска, сортировки и выбранных
параметров фильтрации
*/
func UserLinks(db *pgxpool.Pool,
	userId int,
	search string,
	sortBy string,
	orderBy string,
	filterPeriod string,
	filterViews string,
	filterTags string) ([]models.Link, error) {

	var searchQuery = ""
	if strings.TrimSpace(search) != "" {
		searchQuery = "AND tag ILIKE $2"
	}

	var sortColumn string
	switch sortBy {
	case "date":
		sortColumn = "created_at"
	case "views":
		sortColumn = "views"
	case "tag":
		sortColumn = `LOWER(tag) COLLATE "ru-x-icu"`
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

	var filterPeriodQuery string
	switch filterPeriod {
	case "all":
		filterPeriodQuery = ""
	case "week":
		filterPeriodQuery = "AND created_at >= NOW() - INTERVAL '7 days'"
	case "month":
		filterPeriodQuery = "AND created_at >= NOW() - INTERVAL '1 month'"
	case "year":
		filterPeriodQuery = "AND created_at >= NOW() - INTERVAL '1 year'"
	default:
		return nil, errors.New("Неверный параметр фильтрации по периоду")
	}

	var filterViewsQuery string
	switch filterViews {
	case "0":
		filterViewsQuery = ""
	case "10":
		filterViewsQuery = "AND views >= 10"
	case "100":
		filterViewsQuery = "AND views >= 100"
	case "1000":
		filterViewsQuery = "AND views >= 1000"
	case "10000":
		filterViewsQuery = "AND views >= 10000"
	default:
		return nil, errors.New("Неверный параметр фильтрации по просмотрам")
	}

	var filterTagsQuery string
	switch filterTags {
	case "all":
		filterTagsQuery = ""
	case "with":
		filterTagsQuery = "AND tag <> ''"
	case "without":
		filterTagsQuery = "AND tag = ''"
	default:
		return nil, errors.New("Неверный параметр фильтрации по тэгам")
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
		%s
		%s
		%s
		%s
		ORDER BY %s %s
	`, filterPeriodQuery,
		filterViewsQuery,
		filterTagsQuery,
		searchQuery,
		sortColumn, sortOrder)

	var rows pgx.Rows
	var err error
	if searchQuery == "" {
		rows, err = db.Query(
			context.Background(),
			query,
			userId,
		)
	} else {
		rows, err = db.Query(
			context.Background(),
			query,
			userId,
			"%"+search+"%",
		)
	}

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

/*
IncreaseLinkViews увеличивает количество
просмотров короткой ссылки на единицу
*/
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

/*
DeleteLink удаляет указанную ссылку пользователя
из базы данных
*/
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

/*
AddTag добавляет или обновляет тег
для указанной ссылки
*/
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

/*
CreateCustomLink создаёт пользовательскую короткую
ссылку, если выбранное название ещё не занято,
и возвращает результат создания
*/
func CreateCustomLink(db *pgxpool.Pool, userID int, originalLink string, custom string) (string, bool, error) {
	var link = "localhost:8080/" + custom
	var exist bool

	err := db.QueryRow(
		context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM links
			WHERE short_url=$1
		)`, link).Scan(&exist)

	if err != nil {
		return "", true, err
	}

	if exist {
		return "", true, nil
	}

	_, err = db.Exec(
		context.Background(),
		`INSERT INTO links 
			(user_id, original_url, short_url, is_custom) VALUES ($1, $2, $3, $4)`,
		userID, originalLink, link, true)

	if err != nil {
		return "", true, err
	}

	return link, false, nil
}
