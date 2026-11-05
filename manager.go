package sessions

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Manager struct {
	store Store
	ttl   time.Duration
}

func NewManager(store Store, ttl time.Duration) (*Manager, error) {
	if ttl <= 0 {
		return nil, ErrInvalidTTL
	}
	return &Manager{
		store: store,
		ttl:   ttl,
	}, nil
}

func (m *Manager) Create(ctx context.Context, userID int64) (Session, error) {
	token, err := generateToken()
	if err != nil {
		return Session{}, fmt.Errorf("failed to generate token: %w", err)
	}

	now := time.Now()

	session := Session{
		Token:     token,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(m.ttl),
	}

	stored := session
	stored.Token = hashToken(session.Token)

	if err := m.store.Put(ctx, stored); err != nil {
		return Session{}, fmt.Errorf("failed to put session: %w", err)
	}

	session.Token = token
	return session, nil
}

func (m *Manager) Get(ctx context.Context, token string) (Session, error) {
	session, err := m.store.Get(ctx, hashToken(token))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Session{}, ErrNotFound
		}
		return Session{}, err
	}
	if session.Expired(time.Now()) {
		return Session{}, ErrNotFound
	}
	return session, nil
}

func (m *Manager) GetUser(ctx context.Context, userID int64) ([]Session, error) {
	sessions, err := m.store.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	active := make([]Session, 0, len(sessions))
	for _, s := range sessions {
		if !s.Expired(now) {
			s.Token = ""
			active = append(active, s)
		}
	}
	return active, nil
}

func (m *Manager) Revoke(ctx context.Context, token string) error {
	return m.store.Delete(ctx, hashToken(token))
}

func (m *Manager) RevokeUser(ctx context.Context, userID int64) error {
	return m.store.DeleteUser(ctx, userID)
}

func (m *Manager) Cleanup(ctx context.Context) error {
	return m.store.Cleanup(ctx, time.Now())
}
