package entity

import "time"

type Message struct {
	ID        int
	Author    string
	Text      string
	CreatedAt time.Time
}
