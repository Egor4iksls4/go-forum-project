package entity

import "time"

type Comment struct {
	ID        int
	PostID    int
	Content   string
	Author    string
	CreatedAt time.Time
}
