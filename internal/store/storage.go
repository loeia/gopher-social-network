package store

import "database/sql"

type Post interface {
}
type Storage struct {
	Post
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		NewPostStore(db),
	}
}
