package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Файл database.go отвечает за подключение к PostgreSQL
и инициализацию схемы. Устанавливает пул соединений
через pgxpool и создаёт таблицы links и users, если
их ещё нет в базе
*/

/*
Connect устанавливает пул соединений с PostgreSQL
по строке подключения и проверяет его через Ping.
При ошибке подключения завершает программу через log.Fatal,
так как без БД приложение не может работать
*/
func Connect(databaseURL string) *pgxpool.Pool {

	db, err := pgxpool.New(
		context.Background(),
		databaseURL,
	)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	return db
}

/*
CreateTables создаёт таблицы links и users, если они ещё
не существуют. Безопасно вызывать при каждом запуске сервера,
существующие таблицы не затрагиваются (CREATE TABLE IF NOT EXISTS)
*/
func CreateTables(db *pgxpool.Pool) error {

	_, err := db.Exec(
		context.Background(),
		`
		CREATE TABLE IF NOT EXISTS links (
			id SERIAL PRIMARY KEY,
			original_url TEXT NOT NULL,
			short_url VARCHAR(30) UNIQUE NOT NULL,
			user_id INTEGER DEFAULT 0,
			views INTEGER DEFAULT 0,
			tag TEXT DEFAULT '',
			is_custom BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS users (
		    id SERIAL PRIMARY KEY,
		    username TEXT NOT NULL UNIQUE,
		    email TEXT NOT NULL UNIQUE,
		    password TEXT NOT NULL,
		    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
		    verification_code TEXT NOT NULL DEFAULT '',
		    reset_password_code TEXT NOT NULL DEFAULT '',
		    last_send TIMESTAMPTZ,
		    reset_password_attempts INT DEFAULT 5,
		    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		`,
	)
	return err

}
