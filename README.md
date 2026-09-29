# sessions

A small Go package for creating, storing, and revoking user sessions. It includes an in-memory store, a SQLite store, and an optional `net/http` adapter.

## Install

```sh
go get github.com/nyan2d/sessions
```

## Usage

```go
store := sessions.NewMemoryStore()

manager, err := sessions.NewManager(store, 24*time.Hour)
if err != nil {
    return err
}

session, err := manager.Create(ctx, userID)
if err != nil {
    return err
}

session, err = manager.Get(ctx, session.Token)
if err != nil {
    return err
}
```

- `Get` returns `sessions.ErrNotFound` for a missing or expired session.
- Use `GetUser` to list a user's active sessions, `Revoke` or `RevokeUser` to delete them, and `Cleanup` to remove expired records.

For persistent storage, pass an open `*sql.DB` to `sessions.NewSQLiteStore`. It creates the required table and indexes. The in-memory store does not persist across restarts.

## HTTP adapter

The `github.com/nyan2d/sessions/sessionhttp` package provides cookie helpers and middleware:

```go
adapter, err := sessionhttp.New(manager)
if err != nil {
    return err
}

http.Handle("/account", adapter.Middleware(accountHandler))
```

The middleware rejects missing or invalid sessions. In the handler, use `sessionhttp.Session(r.Context())` to retrieve the session. Use `adapter.SetCookie(w, session)` after creating a session and `adapter.ClearCookie(w)` when signing out.

Cookies are named `session` and are `Secure` by default. Use `sessionhttp.WithCookieName` or `sessionhttp.WithSecure` to change those settings.