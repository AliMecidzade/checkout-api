package domain

type Page struct {
	Limit  int
	Offset int
	Cursor *Cursor
}

type Cursor struct {
	ID int64
}
