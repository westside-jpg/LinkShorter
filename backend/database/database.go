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
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
		`,
	)
	return err

}
