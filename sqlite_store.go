package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {
	sq := &SQLiteStore{
		db: db,
	}

	if err := sq.initTable(); err != nil {
		return nil, err
	}
	return sq, nil
}

func (sq *SQLiteStore) initTable() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token TEXT UNIQUE NOT NULL,
			user_id INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at)`,
	}

	for _, q := range stmts {
		if _, err := sq.db.Exec(q); err != nil {
			return fmt.Errorf("failed to init sqlite schema: %w", err)
		}
	}
	return nil
}

func (sq *SQLiteStore) Put(ctx context.Context, s Session) error {
	q := `INSERT INTO sessions (token, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`
	if _, err := sq.db.ExecContext(ctx, q, s.Token, s.UserID, s.CreatedAt.Unix(), s.ExpiresAt.Unix()); err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}
	return nil
}

func (sq *SQLiteStore) Get(ctx context.Context, token string) (Session, error) {
	q := `SELECT token, user_id, created_at, expires_at FROM sessions WHERE token = ?`
	s, err := scanSession(sq.db.QueryRowContext(ctx, q, token).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("failed to get session: %w", err)
	}
	return s, nil
}

func (sq *SQLiteStore) GetUser(ctx context.Context, userID int64) ([]Session, error) {
	q := `SELECT token, user_id, created_at, expires_at FROM sessions WHERE user_id = ?`
	rows, err := sq.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		session, err := scanSession(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate sessions: %w", err)
	}
	return sessions, nil
}

func (sq *SQLiteStore) Delete(ctx context.Context, token string) error {
	q := `DELETE FROM sessions WHERE token = ?`
	if _, err := sq.db.ExecContext(ctx, q, token); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (sq *SQLiteStore) DeleteUser(ctx context.Context, userID int64) error {
	q := `DELETE FROM sessions WHERE user_id = ?`
	if _, err := sq.db.ExecContext(ctx, q, userID); err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil

}

func (sq *SQLiteStore) Cleanup(ctx context.Context, now time.Time) error {
	nowUnix := now.Unix()
	q := `DELETE FROM sessions WHERE ? >= expires_at`
	if _, err := sq.db.ExecContext(ctx, q, nowUnix); err != nil {
		return fmt.Errorf("failed to cleanup sessions: %w", err)
	}
	return nil
}
