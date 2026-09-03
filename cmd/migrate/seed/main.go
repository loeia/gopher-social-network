package main

import (
	"github.com/loeia/gopherSocialNetwork/internal/db"
	"github.com/loeia/gopherSocialNetwork/internal/env"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

func main() {
	dsn := env.GetString("DB_DSN", "postgres://admin:admin123@localhost/gopher-social-network?sslmode=disable")
	conn, err := db.New(dsn, 3, 3, "15m", "5m")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	store := store.NewStorage(conn)
	db.Seed(store, conn)
}
