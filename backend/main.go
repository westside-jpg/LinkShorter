package main

import (
	"LinkShorter/config"
	"LinkShorter/database"
	"LinkShorter/routers"
	"LinkShorter/services"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := database.Connect(config.GetDatabaseURL())
	err := database.CreateTables(db)
	if err != nil {
		log.Fatal("Не удалось создать таблицу!")
	}
	defer db.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

	routers.SetupRoutes(r, db)

	runDeletion := func() {
		emails, err := services.ScheduleDeletingNotVerifiedUsers(db)
		if err != nil {
			log.Println("Ошибка удаления:", err)
		}
		for _, email := range emails {
			if err := services.SendAccountDeletingEmail(email); err != nil {
				log.Println("Не удалось отправить письмо:", email, err)
			}
		}
	}

	// Горутина для удаления пользователей каждый целый час (9:00, 17:00, 21:00, ...)
	go func() {
		now := time.Now()
		nextHour := now.Truncate(time.Hour).Add(time.Hour)
		log.Println("Горутина начала работать")
		time.Sleep(time.Until(nextHour))

		runDeletion()

		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			runDeletion()
		}
	}()

	err = r.Run()
	if err != nil {
		log.Fatal(err)
	}
}
