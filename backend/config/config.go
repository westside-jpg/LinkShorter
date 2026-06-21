package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetDatabaseURL() string {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Ошибка загрузки .env")
	}

	return os.Getenv("DATABASE_URL")
}
