package sessions

import (
	"context"
	"time"
)

type Store interface {
	Put(ctx context.Context, s Session) error
	Get(ctx context.Context, token string) (Session, error)
	GetUser(ctx context.Context, userID int64) ([]Session, error)
	Delete(ctx context.Context, token string) error
	DeleteUser(ctx context.Context, userID int64) error
	Cleanup(ctx context.Context, now time.Time) error
}
