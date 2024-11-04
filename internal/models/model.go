package models

type Message struct {
	Sender string    `json:"sender"`
	Date   string    `json:"date"`
	Text   string    `json:"text"`
}
