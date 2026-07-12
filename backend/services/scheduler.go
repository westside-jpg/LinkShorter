package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Файл scheduler.go содержит функцию для периодической
очистки базы данных от неверифицированных аккаунтов.
Вызывается по расписанию из фоновой горутины в main.go
раз в ровный час (12:00, 17:00, ...), удаляет пользователей, не подтвердивших почту
в течение трёх дней, вместе с их ссылками
*/

/*
ScheduleDeletingNotVerifiedUsers находит пользователей,
не подтвердивших адрес электронной почты в течение
трёх дней, удаляет их аккаунты и связанные с ними ссылки,
а также возвращает список их адресов электронной почты
*/
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
		`DELETE FROM links
			WHERE user_id IN (
			    SELECT id
				FROM users
				WHERE is_verified = FALSE
				AND created_at < NOW() - INTERVAL '3 days'
			)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(context.Background(),
		`DELETE FROM users
         WHERE is_verified = FALSE
         AND created_at < NOW() - INTERVAL '3 days'
         `)
	if err != nil {
		return nil, err
	}

	return emails, nil
}
