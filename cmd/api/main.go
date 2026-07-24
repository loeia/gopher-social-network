package main

import (
	"log"
	"time"

	"github.com/loeia/gopherSocialNetwork/internal/db"
	"github.com/loeia/gopherSocialNetwork/internal/env"
	"github.com/loeia/gopherSocialNetwork/internal/mailer"
	"github.com/loeia/gopherSocialNetwork/internal/store"
	"go.uber.org/zap"
)

func main() {
	config := config{
		addr:        env.GetString("ADDR", ":8080"),
		env:         env.GetString("ENV", "development"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		db: dbConfig{
			dsn:          env.GetString("DB_DSN", "postgres://admin:admin123@localhost/gopher-social-network?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
			maxLifeTime:  env.GetString("DB_MAX_LIFE_TIME", "5m"),
		},
		mail: mailConfig{
			exp:       time.Hour * 24 * 2, // 2 days
			fromEmail: env.GetString("FROM_EMAIL", ""),
			mailTrap: mailTrapConfig{
				apiKey:      env.GetString("MAILTRAP_API_KEY", ""),
				sandboxUser: env.GetString("MAILTRAP_SANDBOX_USER", ""),
				sandboxPass: env.GetString("MAILTRAP_SANDBOX_PASS", ""),
			},
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

	mailtrap, err := mailer.NewMailTrapClient(
		config.mail.fromEmail,
		config.mail.mailTrap.apiKey,
		config.mail.mailTrap.sandboxUser,
		config.mail.mailTrap.sandboxPass,
	)
	if err != nil {
		log.Fatalln(err)
	}

	app := &application{
		config: config,
		store:  store,
		logger: logger,
		mailer: mailtrap,
	}

	logger.Fatalln(app.run(app.mount()))
}
