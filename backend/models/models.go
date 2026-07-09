package models

import "time"

type Link struct {
	LinkID      string    `json:"id"`
	ShortURL    string    `json:"short"`
	OriginalURL string    `json:"original"`
	Views       int       `json:"views"`
	CreatedAt   time.Time `json:"created_at"`
}
