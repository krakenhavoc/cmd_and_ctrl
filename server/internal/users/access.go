package users

// access.go is the user store's half of ADR 0156 §1: the last answer
// the sign-up gate got for a person (users.access_checked_at and
// users.access_result, migration 0019), and writing back the rotated
// refresh token a renewal's re-check receives.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// ErrInvalidAccess is RecordAccessCheck refusing an answer that is not
// one of discord.AccessAdmitted, AccessNotMember or AccessNoRole.
var ErrInvalidAccess = errors.New("users: invalid access result")

// RecordAccessCheck stores Discord's answer for the user, and when it
// was given. Only an answer is recorded: an outage is not one, and the
// caller does not call this for it. ErrNotFound for an unknown user.
func (s *SQLStore) RecordAccessCheck(ctx context.Context, id uuid.UUID, result discord.Access, at time.Time) error {
	switch result {
	case discord.AccessAdmitted, discord.AccessNotMember, discord.AccessNoRole:
	default:
		return fmt.Errorf("%w: %q", ErrInvalidAccess, result)
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET access_checked_at = ?, access_result = ? WHERE id = ?`,
		at.UnixMilli(), string(result), id.String())
	if err != nil {
		return fmt.Errorf("users: record access check for %s: %w", id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// AccessCheck reads the last recorded answer: zero time and "" when the
// person has never been checked.
func (s *SQLStore) AccessCheck(ctx context.Context, id uuid.UUID) (time.Time, discord.Access, error) {
	var (
		at     sql.NullInt64
		result sql.NullString
	)
	err := s.db.QueryRowContext(ctx, `SELECT access_checked_at, access_result FROM users WHERE id = ?`, id.String()).Scan(&at, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, "", ErrNotFound
	}
	if err != nil {
		return time.Time{}, "", fmt.Errorf("users: access check for %s: %w", id, err)
	}
	if !at.Valid {
		return time.Time{}, discord.Access(result.String), nil
	}
	return time.UnixMilli(at.Int64).UTC(), discord.Access(result.String), nil
}

// SetRefreshToken replaces the stored refresh token of a Discord
// identity with a rotated one, sealed like UpsertFromDiscord seals it.
// With no key configured nothing is stored (and RefreshToken never
// returned one to rotate), so it is a no-op. An empty token is a no-op
// too: Discord returning none must not erase the one that still works.
// ErrNotFound when there is no such identity.
func (s *SQLStore) SetRefreshToken(ctx context.Context, subject, token string) error {
	if s.sealer == nil || token == "" {
		return nil
	}
	sealed, err := s.sealer.Seal([]byte(token), refreshTokenAD(ProviderDiscord, subject))
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE identities SET refresh_token = ? WHERE provider = ? AND subject = ?`,
		sealed, ProviderDiscord, subject)
	if err != nil {
		return fmt.Errorf("users: store rotated refresh token: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
