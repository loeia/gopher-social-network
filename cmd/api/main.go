package main

import (
	"log"
	"time"

	"github.com/loeia/gopherSocialNetwork/internal/auth"
	"github.com/loeia/gopherSocialNetwork/internal/db"
	"github.com/loeia/gopherSocialNetwork/internal/env"
	"github.com/loeia/gopherSocialNetwork/internal/mailer"
	"github.com/loeia/gopherSocialNetwork/internal/ratelimiter"
	"github.com/loeia/gopherSocialNetwork/internal/store"
	"github.com/loeia/gopherSocialNetwork/internal/store/cache"
	"github.com/redis/go-redis/v9"
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
			resetExp:  time.Minute * 15,
			fromEmail: env.GetString("FROM_EMAIL", ""),
			mailTrap: mailTrapConfig{
				apiKey:      env.GetString("MAILTRAP_API_KEY", ""),
				sandboxUser: env.GetString("MAILTRAP_SANDBOX_USER", ""),
				sandboxPass: env.GetString("MAILTRAP_SANDBOX_PASS", ""),
			},
		},
		auth: authConfig{
			token: tokenConfig{
				secret: env.GetString("AUTH_SECRET", ""),
				exp:    time.Hour * time.Duration(env.GetInt("AUTH_EXP", 72)), // default 3 days
				iss:    env.GetString("AUTH_ISSUER", "Bearer"),
			},
		},
		redisCfg: redisConfig{
			addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			password: env.GetString("REDIS_PASSWORD", ""),
			db:       env.GetInt("REDIS_DB", 0),
			enabled:  env.GetBool("REDIS_ENABLED", false),
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS_COUNT", 20),
			TimeFrame:            time.Duration(env.GetInt("RATELIMITER_TIME_FRAME", 5)) * time.Second,
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", true),
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

	// cache
	var rdb *redis.Client
	if config.redisCfg.enabled {
		rdb = cache.NewRedisClient(config.redisCfg.addr, config.redisCfg.password, config.redisCfg.db)
		logger.Info("redis cache connection established!")
	}
	cacheStore := cache.NewCacheStorage(rdb)

	// email
	mailtrap, err := mailer.NewMailTrapClient(
		config.mail.fromEmail,
		config.mail.mailTrap.apiKey,
		config.mail.mailTrap.sandboxUser,
		config.mail.mailTrap.sandboxPass,
	)
	if err != nil {
		log.Fatalln(err)
	}

	// authenticator
	jwtAuthenticator := auth.NewJWTAuthenticator(config.auth.token.secret, config.auth.token.iss, config.auth.token.iss)

	// rate limiter
	var rl ratelimiter.Limiter
	if config.rateLimiter.Enabled && config.redisCfg.enabled {
		rl = ratelimiter.NewRedisRateLimiter(rdb, config.rateLimiter.RequestsPerTimeFrame, config.rateLimiter.TimeFrame)
		logger.Info("redis rate limiter enabled!")
	} else {
		config.rateLimiter.Enabled = false
	}

	app := &application{
		config:        config,
		store:         store,
		cache:         cacheStore,
		logger:        logger,
		mailer:        mailtrap,
		authenticator: jwtAuthenticator,
		rateLimiter:   rl,
	}

	logger.Fatalln(app.run(app.mount()))
}
