package main

import (
	"LinkShorter/config"
	"LinkShorter/database"
	"LinkShorter/routers"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	/*
	   Структура БД в PostgreSQL:
	   id, code, original_url, short_url, created_at
	*/

	db := database.Connect(config.GetDatabaseURL())
	err := database.CreateTables(db)
	if err != nil {
		log.Fatal("Не удалось создать таблицу!")
	}
	defer db.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Content-Type"},
	}))

	routers.SetupRoutes(r, db)

	err = r.Run()
	if err != nil {
		log.Fatal(err)
	}
}
