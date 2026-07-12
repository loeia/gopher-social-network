package store

import "database/sql"

type PostStore struct {
	DB *sql.DB
}

func NewPostStore(db *sql.DB) *PostStore {
	return &PostStore{
		DB: db,
	}
}
