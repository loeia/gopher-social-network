package store

import (
	"context"
	"database/sql"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

type RoelStore struct {
	db *sql.DB
}

func NewRoleStore(db *sql.DB) *RoelStore {
	return &RoelStore{db}
}

func (s *RoelStore) GetByName(c context.Context, roleName string) (*Role, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := "select id,name,description,level FROM roles WHERE name = $1"

	role := new(Role)
	if err := s.db.QueryRowContext(ctx, query, roleName).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Level,
	); err != nil {
		return nil, err
	}

	return role, nil
}
