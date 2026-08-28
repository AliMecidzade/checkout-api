package domain

import "time"

type Page struct {
	Limit  int
	Offset int
	Cursor *Cursor
}

type Cursor struct {
	CreatedAt time.Time
	ID        int64
}
