package sessions

import (
	"context"
	"sync"
	"time"
)

type MemoryStore struct {
	store sync.Map
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (m *MemoryStore) Put(ctx context.Context, s Session) error {
	m.store.Store(s.Token, s)
	return nil
}

func (m *MemoryStore) Get(ctx context.Context, token string) (Session, error) {
	if val, ok := m.store.Load(token); ok {
		return val.(Session), nil
	}
	return Session{}, ErrNotFound
}

func (m *MemoryStore) GetUser(ctx context.Context, userID int64) ([]Session, error) {
	result := []Session{}

	m.store.Range(func(key, value any) bool {
		if value.(Session).UserID == userID {
			result = append(result, value.(Session))
		}
		return true
	})

	return result, nil
}

func (m *MemoryStore) Delete(ctx context.Context, token string) error {
	m.store.Delete(token)
	return nil
}

func (m *MemoryStore) DeleteUser(ctx context.Context, userID int64) error {
	m.store.Range(func(key, value any) bool {
		if value.(Session).UserID == userID {
			m.store.Delete(key)
		}
		return true
	})
	return nil
}

func (m *MemoryStore) Cleanup(ctx context.Context, now time.Time) error {
	m.store.Range(func(key, value any) bool {
		s := value.(Session)
		if s.Expired(now) {
			m.store.Delete(key)
		}
		return true
	})
	return nil
}
