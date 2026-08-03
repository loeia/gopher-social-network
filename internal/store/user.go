package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	IsActive  bool     `json:"is_active"`
	RoleID    int      `json:"role_id"`
	Role      Role     `json:"role"`
	TokenVer  int      `json:"token_ver"`
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
func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
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
	query := `
		INSERT INTO users (username,password,email,role_id)
		VALUES ($1,$2,$3,(SELECT id FROM roles WHERE name = $4))
		RETURNING id,created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if user.Role.Name == "" {
		user.Role.Name = "user"
	}

	row := tx.QueryRowContext(ctx, query, user.Username, user.Password.hash, user.Email, user.Role.Name)
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

	query := `
			SELECT u.id,u.username,u.email,u.password,u.created_at,u.is_active,u.token_ver,
				   r.id,r.name,r.description,r.level
			FROM users u
			JOIN roles r ON r.id = u.role_id
			WHERE u.id = $1 AND u.is_active = true
	`

	var user User
	if err := s.db.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.IsActive,
		&user.TokenVer,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Description,
		&user.Role.Level,
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
	query := `INSERT INTO user_invitations (user_id,token,expiry) VALUES ($1,$2,$3)`

	_, err := tx.ExecContext(c, query, userId, token, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Activate(c context.Context, token string) error {
	return withTx(s.db, c, func(tx *sql.Tx) error {
		ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
		defer cancel()

		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id,u.username ,u.email,u.created_at,u.is_active FROM users u
		JOIN user_invitations ui ON ui.user_id = u.id
		WHERE ui.token = $1 AND ui.expiry >= $2
	`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	user := User{}
	if err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	if _, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *UserStore) deleteUserInvitations(c context.Context, tx *sql.Tx, userId int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Delete(c context.Context, userId int64) error {
	return withTx(s.db, c, func(tx *sql.Tx) error {
		ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
		defer cancel()

		if err := s.delete(ctx, tx, userId); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(c, tx, userId); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) delete(c context.Context, tx *sql.Tx, userId int64) error {
	query := `DELETE FROM users WHERE id = $1`

	if _, err := tx.ExecContext(c, query, userId); err != nil {
		return err
	}

	return nil
}

func (s *UserStore) GetByEmail(c context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := "SELECT id,email,username,password,created_at,token_ver FROM users WHERE email = $1 AND is_active = true"

	var user User
	if err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password.hash,
		&user.CreatedAt,
		&user.TokenVer,
	); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil

}

func (s *UserStore) UpdatePassword(c context.Context, newPass string, userId int64) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	var p password
	if err := p.Set(newPass); err != nil {
		return err
	}

	query := "UPDATE users SET password = $1,token_ver = token_ver + 1 WHERE id = $2"

	if _, err := s.db.ExecContext(ctx, query, p.hash, userId); err != nil {
		return err
	}

	return nil
}
