package entity

import "time"

type Message struct {
	ID        int
	Author    int
	Text      string
	CreatedAt time.Time
}
