package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
