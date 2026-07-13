package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetDatabaseURL() string {
	err := godotenv.Load()

	if err != nil {
		log.Println(".env")
	}

	return os.Getenv("DATABASE_URL")
}
