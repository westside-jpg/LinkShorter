package models

import "time"

type Link struct {
	ID          int
	OriginalURL string
	ShortURL    string
	CreatedAt   time.Time
}
