package store

import "context"

func (s *UserStore) BanByUsername(c context.Context, username string) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `UPDATE users SET is_active = false WHERE username = $1`

	_, err := s.db.ExecContext(ctx, query, username)

	return err
}

func (s *UserStore) UnbanByUsername(c context.Context, username string) error {
	ctx, cancel := context.WithTimeout(c, QueryTimeoutDuration)
	defer cancel()

	query := `UPDATE users SET is_active = true WHERE username = $1`

	_, err := s.db.ExecContext(ctx, query, username)

	return err
}
