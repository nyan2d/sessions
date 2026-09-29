package sessionhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/nyan2d/sessions"
)

type contextKey struct{}

type Adapter struct {
	manager    *sessions.Manager
	cookieName string
	secure     bool
}

type Option func(*Adapter)

func WithCookieName(name string) Option {
	return func(a *Adapter) { a.cookieName = name }
}

func WithSecure(secure bool) Option {
	return func(a *Adapter) { a.secure = secure }
}

func New(manager *sessions.Manager, opts ...Option) (*Adapter, error) {
	if manager == nil {
		return nil, errors.New("sessionhttp: nil manager")
	}

	a := &Adapter{
		manager:    manager,
		cookieName: "session",
		secure:     true,
	}
	for _, opt := range opts {
		opt(a)
	}

	if (&http.Cookie{Name: a.cookieName, Value: "x"}).String() == "" {
		return nil, errors.New("sessionhttp: invalid cookie name")
	}
	return a, nil
}

func (a *Adapter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(a.cookieName)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		session, err := a.manager.Get(r.Context(), cookie.Value)
		switch {
		case errors.Is(err, sessions.ErrNotFound):
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		case err != nil:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), contextKey{}, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Session(ctx context.Context) (sessions.Session, bool) {
	s, ok := ctx.Value(contextKey{}).(sessions.Session)
	return s, ok
}

func (a *Adapter) SetCookie(w http.ResponseWriter, s sessions.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cookieName,
		Value:    s.Token,
		Path:     "/",
		Expires:  s.ExpiresAt,
		HttpOnly: true,
		Secure:   a.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *Adapter) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.secure,
		SameSite: http.SameSiteLaxMode,
	})
}
