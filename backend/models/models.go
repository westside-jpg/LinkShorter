package models

type Link struct {
	LinkID      string `json:"id"`
	ShortURL    string `json:"short"`
	OriginalURL string `json:"original"`
	Views       int    `json:"views"`
}
