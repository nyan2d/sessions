package sessions

import (
	"fmt"
	"time"
)

type Session struct {
	Token     string    `json:"-"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *Session) Expired(t time.Time) bool {
	return !t.Before(s.ExpiresAt)
}

func scanSession(scanf func(dest ...any) error) (Session, error) {
	var (
		s                    Session
		createdAt, expiresAt int64
	)

	if err := scanf(&s.Token, &s.UserID, &createdAt, &expiresAt); err != nil {
		return Session{}, fmt.Errorf("failed to scan session: %w", err)
	}
	s.CreatedAt = time.Unix(createdAt, 0)
	s.ExpiresAt = time.Unix(expiresAt, 0)

	return s, nil
}
