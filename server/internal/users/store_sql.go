package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// SQLStore is the users / identities tables from migration 0003.
// Hand-written SQL over database/sql, as internal/db asks. Every *_at
// column is Unix milliseconds.
type SQLStore struct {
	db     *sql.DB
	sealer *Sealer // nil: no CMDCTRL_IDENTITY_KEY, refresh tokens are discarded
	now    func() time.Time
}

// NewSQLStore wraps an opened, migrated database. sealer may be nil,
// in which case no refresh token is ever written (the column is set
// to NULL). The caller owns d and closes it.
func NewSQLStore(d *db.DB, sealer *Sealer) *SQLStore {
	return &SQLStore{db: d.DB, sealer: sealer, now: func() time.Time { return time.Now().UTC() }}
}

// refreshTokenAD is the associated data a refresh token is sealed
// with: the identity it belongs to. A sealed blob moved to another
// row fails to open.
func refreshTokenAD(provider, subject string) []byte {
	return []byte(provider + "\x00" + subject)
}

// UpsertFromDiscord implements Store. See the interface for the
// contract.
//
// It runs in one transaction. Two sign-ins for the same snowflake
// racing each other (two tabs, one click each) can both find no
// identity and both try to insert one, or both read and then both try
// to write; the loser gets a constraint failure or SQLITE_BUSY and is
// retried once, when it finds the winner's row and takes the update
// path.
func (s *SQLStore) UpsertFromDiscord(ctx context.Context, profile discord.User, refreshToken, scopes string) (User, error) {
	if profile.ID == "" {
		return User{}, errors.New("users: Discord profile has no id")
	}
	u, err := s.upsertDiscord(ctx, profile, refreshToken, scopes)
	if err != nil && (isConstraint(err) || isBusy(err)) {
		u, err = s.upsertDiscord(ctx, profile, refreshToken, scopes)
	}
	return u, err
}

func (s *SQLStore) upsertDiscord(ctx context.Context, profile discord.User, refreshToken, scopes string) (User, error) {
	now := s.now().Truncate(time.Millisecond)
	nowMs := now.UnixMilli()
	name := profile.DisplayName()
	avatar := AvatarPath(profile.ID, profile.Avatar)

	// sealed is the refresh_token column's new value. keep says
	// "leave the stored one alone" (a sign-in that returned no token,
	// with a key configured). With no key the column is always
	// written NULL: a token is never stored unencrypted, and a
	// ciphertext from when a key was configured is not left behind
	// to be opened by a key that is no longer the configured one.
	var (
		sealed []byte
		keep   bool
	)
	switch {
	case s.sealer == nil:
		// discard
	case refreshToken == "":
		keep = true
	default:
		var err error
		sealed, err = s.sealer.Seal([]byte(refreshToken), refreshTokenAD(ProviderDiscord, profile.ID))
		if err != nil {
			return User{}, err
		}
	}
	// A nil []byte is not reliably bound as NULL by every driver; an
	// untyped nil is.
	var tokenCol any
	if sealed != nil {
		tokenCol = sealed
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var userID string
	err = tx.QueryRowContext(ctx,
		`SELECT user_id FROM identities WHERE provider = ? AND subject = ?`,
		ProviderDiscord, profile.ID).Scan(&userID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		userID = uuid.New().String()
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO users (id, display_name, avatar_url, created_at, last_seen_at) VALUES (?, ?, ?, ?, ?)`,
			userID, name, nullString(avatar), nowMs, nowMs); err != nil {
			return User{}, fmt.Errorf("users: insert user: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO identities (provider, subject, user_id, display_name, avatar_hash, refresh_token, scopes, linked_at, refreshed_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			ProviderDiscord, profile.ID, userID, name, nullString(profile.Avatar), tokenCol, scopes, nowMs, nowMs); err != nil {
			return User{}, fmt.Errorf("users: insert identity: %w", err)
		}
	case err != nil:
		return User{}, fmt.Errorf("users: find identity: %w", err)
	default:
		if _, err := tx.ExecContext(ctx,
			`UPDATE users SET display_name = ?, avatar_url = ?, last_seen_at = ? WHERE id = ?`,
			name, nullString(avatar), nowMs, userID); err != nil {
			return User{}, fmt.Errorf("users: update user: %w", err)
		}
		q := `UPDATE identities SET display_name = ?, avatar_hash = ?, scopes = ?, refreshed_at = ?, refresh_token = ?
		      WHERE provider = ? AND subject = ?`
		args := []any{name, nullString(profile.Avatar), scopes, nowMs, tokenCol, ProviderDiscord, profile.ID}
		if keep {
			q = `UPDATE identities SET display_name = ?, avatar_hash = ?, scopes = ?, refreshed_at = ?
			     WHERE provider = ? AND subject = ?`
			args = []any{name, nullString(profile.Avatar), scopes, nowMs, ProviderDiscord, profile.ID}
		}
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			return User{}, fmt.Errorf("users: update identity: %w", err)
		}
	}

	u, err := getUser(ctx, tx, userID)
	if err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return u, nil
}

// Get implements Store.
func (s *SQLStore) Get(ctx context.Context, id uuid.UUID) (User, error) {
	return getUser(ctx, s.db, id.String())
}

// RefreshToken opens the stored refresh token for a Discord identity.
// ok is false when none is stored (no key at the time, or never
// returned). An error means a token is stored but cannot be opened —
// most likely CMDCTRL_IDENTITY_KEY changed since it was sealed.
//
// Nothing calls this yet: silent profile refresh (decision 5) is its
// first user. It exists now so the round trip is tested against the
// real column rather than only against Sealer.
func (s *SQLStore) RefreshToken(ctx context.Context, subject string) (token string, ok bool, err error) {
	var sealed []byte
	err = s.db.QueryRowContext(ctx,
		`SELECT refresh_token FROM identities WHERE provider = ? AND subject = ?`,
		ProviderDiscord, subject).Scan(&sealed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, ErrNotFound
	}
	if err != nil {
		return "", false, err
	}
	if sealed == nil {
		return "", false, nil
	}
	if s.sealer == nil {
		return "", false, fmt.Errorf("users: a refresh token is stored but %s is not set", IdentityKeyEnv)
	}
	pt, err := s.sealer.Open(sealed, refreshTokenAD(ProviderDiscord, subject))
	if err != nil {
		return "", false, err
	}
	return string(pt), true, nil
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func getUser(ctx context.Context, q queryRower, id string) (User, error) {
	var (
		u                                User
		idStr                            string
		avatar                           sql.NullString
		created, lastSeen, invalidBefore int64
	)
	err := q.QueryRowContext(ctx,
		`SELECT id, display_name, avatar_url, created_at, last_seen_at, sessions_invalid_before FROM users WHERE id = ?`,
		id).Scan(&idStr, &u.DisplayName, &avatar, &created, &lastSeen, &invalidBefore)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("users: get %s: %w", id, err)
	}
	if u.ID, err = uuid.Parse(idStr); err != nil {
		return User{}, fmt.Errorf("users: stored id %q: %w", idStr, err)
	}
	u.AvatarURL = avatar.String
	u.CreatedAt = time.UnixMilli(created).UTC()
	u.LastSeenAt = time.UnixMilli(lastSeen).UTC()
	if invalidBefore != 0 {
		u.SessionsInvalidBefore = time.UnixMilli(invalidBefore).UTC()
	}
	return u, nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// isConstraint reports a SQLite constraint violation. modernc surfaces
// these as an error whose text carries "constraint failed"; there is
// no exported sentinel to test against.
func isConstraint(err error) bool {
	return err != nil && strings.Contains(err.Error(), "constraint failed")
}

// isBusy reports SQLITE_BUSY (including the WAL snapshot variant a
// read transaction gets when it tries to write after another
// connection already has).
func isBusy(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "SQLITE_BUSY") || strings.Contains(err.Error(), "database is locked"))
}

var _ Store = (*SQLStore)(nil)
var _ Store = NoStore{}
