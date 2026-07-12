package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(dsn string, maxOpenConns, maxIdleConns int, maxIdleTime, maxLifeTime string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	if err := dbConf(db, maxOpenConns, maxIdleConns, maxIdleTime, maxLifeTime); err != nil {
		return nil, err
	}

	return db, nil
}

func dbConf(db *sql.DB, maxOpenConns, maxIdleConns int, maxIdleTime, maxLifeTime string) error {
	idleTime, err := time.ParseDuration(maxIdleTime)
	if err != nil {
		return err
	}
	lifeTime, err := time.ParseDuration(maxLifeTime)
	if err != nil {
		return err
	}
	db.SetConnMaxIdleTime(idleTime)
	db.SetConnMaxLifetime(lifeTime)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	return nil
}
