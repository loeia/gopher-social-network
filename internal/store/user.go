package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (s *UserStore) Create(ctx context.Context, user *User, tx *sql.Tx) error {
	query := `insert into users (username,password,email) values ($1,$2,$3) returning id,created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	row := tx.QueryRowContext(ctx, query, user.Username, user.Password.hash, user.Email)
	err := row.Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch {
			case strings.Contains(pgErr.Message, "users_email_key"):
				return ErrDuplicateEmail
			case strings.Contains(pgErr.Message, "users_username_key"):
				return ErrDuplicateUsername
			}
		}
		return err
	}

	return nil
}

func (s *UserStore) GetById(c context.Context, userId int64) (*User, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := "SELECT id,username,email,password,created_at FROM users WHERE id = $1"

	var user User
	if err := s.db.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) CreateAndInvite(c context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(s.db, c, func(tx *sql.Tx) error {
		ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
		defer cancel()

		if err := s.Create(ctx, user, tx); err != nil {
			return err
		}

		if err := s.createUserInvitation(ctx, tx, invitationExp, token, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) createUserInvitation(c context.Context, tx *sql.Tx, exp time.Duration, token string, userId int64) error {
	query := `INSERT INTO users_invitations (user_id,token,expriy) VALUES ($1,$2,$3)`

	_, err := tx.ExecContext(c, query, userId, token, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}
