package models

type Link struct {
	ShortURL    string `json:"short"`
	OriginalURL string `json:"original"`
	Views       int    `json:"views"`
}
