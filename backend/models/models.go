package models

import "time"

/*
Файл models.go содержит структуру Link, отражающую строку
таблицы links из БД. Используется в services.UserLinks
для чтения ссылок из базы и в роуте /api/my-links при
сериализации ответа в JSON
*/

type Link struct {
	LinkID      int       `json:"id"`
	ShortURL    string    `json:"short"`
	OriginalURL string    `json:"original"`
	Views       int       `json:"views"`
	Tag         string    `json:"tag"`
	CreatedAt   time.Time `json:"created_at"`
}
