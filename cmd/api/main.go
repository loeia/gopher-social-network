package main

import (
	"time"

	"github.com/loeia/gopherSocialNetwork/internal/db"
	"github.com/loeia/gopherSocialNetwork/internal/env"
	"github.com/loeia/gopherSocialNetwork/internal/store"
	"go.uber.org/zap"
)

func main() {
	config := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			dsn:          env.GetString("DB_DSN", "postgres://admin:admin123@localhost/gopher-social-network?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
			maxLifeTime:  env.GetString("DB_MAX_LIFE_TIME", "5m"),
		},
		mail: mailConfig{
			exp: time.Hour * 24 * 2, // 2 days
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(config.db.dsn, config.db.maxOpenConns, config.db.maxIdleConns, config.db.maxIdleTime, config.db.maxLifeTime)
	if err != nil {
		logger.Fatalln(err)
	}
	defer db.Close()
	logger.Info("database connection pool established!")

	store := store.NewStorage(db)

	app := &application{
		config: config,
		store:  store,
		logger: logger,
	}

	logger.Fatalln(app.run(app.mount()))
}
