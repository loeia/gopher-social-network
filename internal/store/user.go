package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Password       password  `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	IsActive       bool      `json:"is_active"`
	RoleID         int       `json:"role_id"`
	Role           Role      `json:"role"`
	TokenVer       int       `json:"token_ver"`
	AvatarURL      string    `json:"avatar_url"`
	Bio            string    `json:"bio"`
	Links          []string  `json:"links"`
	ShowEmail      bool      `json:"show_email"`
	FollowersCount int64     `json:"followers_count"`
	FollowingCount int64     `json:"following_count"`
	PostsCount     int64     `json:"posts_count"`
	LikesCount     int64     `json:"likes_count"`
	RepliesCount   int64     `json:"replies_count"`
}

type UserFollower struct {
	FollowerID int64     `json:"follower_id"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserFollowing struct {
	FollowingID int64     `json:"following_id"`
	Username    string    `json:"username"`
	CreatedAt   time.Time `json:"created_at"`
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

	user.AvatarURL = fmt.Sprintf("/users/%d/avatar", user.ID)

	return nil
}

func (s *UserStore) GetById(c context.Context, userId int64) (*User, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

		query := `
			SELECT
				u.id,u.username,u.email,u.password,u.created_at,u.is_active,u.token_ver,
				COALESCE(u.bio,''),u.links,
				u.show_email,
				u.followers_count,u.following_count,u.posts_count,u.likes_count,u.replies_count,
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
		&user.Bio,
		pq.Array(&user.Links),
		&user.ShowEmail,
		&user.FollowersCount,
		&user.FollowingCount,
		&user.PostsCount,
		&user.LikesCount,
		&user.RepliesCount,
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

	user.AvatarURL = fmt.Sprintf("/users/%d/avatar", user.ID)

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

	user.AvatarURL = fmt.Sprintf("/users/%d/avatar", user.ID)

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

func (s *UserStore) UpdateAvatar(c context.Context, userId int64, data []byte, mime string) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `UPDATE users SET avatar = $1, avatar_mime = $2 WHERE id = $3`

	result, err := s.db.ExecContext(ctx, query, data, mime, userId)
	if err != nil {
		return err
	}

	if _, err := result.RowsAffected(); err != nil {
		return err
	}

	return nil
}

func (s *UserStore) GetAvatar(c context.Context, userId int64) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `SELECT avatar, avatar_mime FROM users WHERE id = $1`

	var data []byte
	var mime sql.NullString
	if err := s.db.QueryRowContext(ctx, query, userId).Scan(&data, &mime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}

	if len(data) == 0 {
		return nil, "", ErrNotFound
	}

	return data, mime.String, nil
}

func (s *UserStore) GetUserFollowing(c context.Context, userId int64, pgq *PaginationQuery) ([]*UserFollowing, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
				SELECT f.user_id,u.username,f.created_at FROM followers f
				LEFT JOIN users u ON u.id = f.user_id
				WHERE f.follower_id = $1
		ORDER BY f.created_at ` + pgq.Sort + `
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, userId, pgq.Limit, pgq.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []*UserFollowing
	for rows.Next() {
		var f UserFollowing
		if err := rows.Scan(&f.FollowingID, &f.Username, &f.CreatedAt); err != nil {
			return nil, err
		}
		followers = append(followers, &f)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return followers, nil

}

func (s *UserStore) GetUserFollowers(c context.Context, userId int64, pgq *PaginationQuery) ([]*UserFollower, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
			SELECT f.follower_id,u.username,f.created_at FROM followers f
			LEFT JOIN users u ON u.id = f.follower_id
			WHERE f.user_id = $1
			ORDER BY f.created_at ` + pgq.Sort + `
			LIMIT $2 OFFSET $3
		`

	rows, err := s.db.QueryContext(ctx, query, userId, pgq.Limit, pgq.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []*UserFollower
	for rows.Next() {
		var f UserFollower
		if err := rows.Scan(&f.FollowerID, &f.Username, &f.CreatedAt); err != nil {
			return nil, err
		}
		followers = append(followers, &f)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return followers, nil
}

func (s *UserStore) GetUserOwnPosts(c context.Context, userId int64, pgq *PaginationQuery) ([]*Post, error) {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT
		    p.id,p.user_id,p.title,p.content,p.tags,p.version,
		    p.created_at,p.updated_at,p.comment_count,p.like_count,p.view_count,
		    u.username
		FROM posts p
		JOIN users u ON u.id = p.user_id
		WHERE p.user_id = $1
		ORDER BY p.created_at ` + pgq.Sort + `
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, userId, pgq.Limit, pgq.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			pq.Array(&p.Tags),
			&p.Version,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.CommentCount,
			&p.LikeCount,
			&p.ViewCount,
			&p.User.Username,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &p)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return posts, nil
}

func (s *UserStore) UpdateProfile(c context.Context, userId int64, bio string, links []string, showEmail bool) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `
		UPDATE users
		SET bio = $1, links = $2, show_email = $3
		WHERE id = $4 AND is_active = true
	`
	if _, err := s.db.ExecContext(ctx, query, bio, links, showEmail, userId); err != nil {
		return err
	}
	return nil
}

func (s *UserStore) CreatePasswordReset(c context.Context, token string, userId int64, expiry time.Duration) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := "INSERT INTO password_resets (token,user_id,expiry) VALUES ($1,$2,$3)"

	_, err := s.db.ExecContext(ctx, query, token, userId, time.Now().Add(expiry))

	return err
}

func (s *UserStore) ResetPassword(c context.Context, token string, newPassword string) error {
	return withTx(s.db, c, func(tx *sql.Tx) error {
		ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
		defer cancel()

		hash := sha256.Sum256([]byte(token))
		hashToken := hex.EncodeToString(hash[:])

		query := `SELECT user_id FROM password_resets WHERE token = $1 AND expiry >= $2`

		var userId int64
		if err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(&userId); err != nil {
			switch err {
			case sql.ErrNoRows:
				return ErrNotFound
			default:
				return err
			}
		}

		var p password
		if err := p.Set(newPassword); err != nil {
			return err
		}

		updateQuery := `UPDATE users SET password = $1, token_ver = token_ver + 1 WHERE id = $2`
		if _, err := tx.ExecContext(ctx, updateQuery, p.hash, userId); err != nil {
			return err
		}

		deleteQuery := `DELETE FROM password_resets WHERE user_id = $1`
		if _, err := tx.ExecContext(ctx, deleteQuery, userId); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) Rename(c context.Context, userId int64, newName string) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `UPDATE users SET username = $1 WHERE id = $2 `

	result, err := s.db.ExecContext(ctx, query, newName, userId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.Message, "users_username_key") {
				return ErrDuplicateUsername
			}
		}
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *UserStore) DeleteUserAvatar(c context.Context, userId int64) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `UPDATE users SET avatar = NULL, avatar_mime = NULL WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, userId)

	return err
}
