package main

import (
	"LinkShorter/config"
	"LinkShorter/database"
	"LinkShorter/routers"
	"log"

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
	routers.SetupRoutes(r, db)
	r.LoadHTMLGlob("../frontend/templates/*")

	err = r.Run()
	if err != nil {
		log.Fatal(err)
	}
}
