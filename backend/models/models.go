package models

import "time"

type Link struct {
	LinkID      int       `json:"id"`
	ShortURL    string    `json:"short"`
	OriginalURL string    `json:"original"`
	Views       int       `json:"views"`
	Tag         string    `json:"tag"`
	CreatedAt   time.Time `json:"created_at"`
}
