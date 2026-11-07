package users

// role_connection.go is the user store's half of ADR 0163 §4: the sealed
// refresh token of a person's Linked Roles grant
// (identities.role_connection_token, migration 0035), kept apart from
// the sign-in grant's refresh_token, which every sign-in replaces.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrNoSealer is a role-connection write on a store with no
// CMDCTRL_IDENTITY_KEY: the token could not be kept, so it is not.
var ErrNoSealer = fmt.Errorf("users: %s is not set, so a role-connection token cannot be stored", IdentityKeyEnv)

// roleConnectionAD is the associated data the role-connection token is
// sealed with: the identity, and which column. A sealed blob moved to
// another row, or between this column and refresh_token, fails to open.
func roleConnectionAD(subject string) []byte {
	return []byte(ProviderDiscord + "\x00" + subject + "\x00role_connection")
}

// CanSeal reports whether the store has a key to seal tokens with.
func (s *SQLStore) CanSeal() bool { return s.sealer != nil }

// SetRoleConnectionToken stores (or replaces) the role-connection refresh
// token of a Discord identity, sealed. An empty token is a no-op, as for
// SetRefreshToken: Discord returning none must not erase one that still
// works. ErrNoSealer with no key; ErrNotFound when there is no such
// identity.
func (s *SQLStore) SetRoleConnectionToken(ctx context.Context, subject, token string) error {
	if token == "" {
		return nil
	}
	if s.sealer == nil {
		return ErrNoSealer
	}
	sealed, err := s.sealer.Seal([]byte(token), roleConnectionAD(subject))
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE identities SET role_connection_token = ? WHERE provider = ? AND subject = ?`,
		sealed, ProviderDiscord, subject)
	if err != nil {
		return fmt.Errorf("users: store role-connection token: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// RoleConnectionToken opens the stored role-connection token. ok is
// false when the person has not linked (NULL). An error means a token is
// stored but cannot be opened: most likely CMDCTRL_IDENTITY_KEY changed.
func (s *SQLStore) RoleConnectionToken(ctx context.Context, subject string) (token string, ok bool, err error) {
	var sealed []byte
	err = s.db.QueryRowContext(ctx,
		`SELECT role_connection_token FROM identities WHERE provider = ? AND subject = ?`,
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
	pt, err := s.OpenRoleConnectionToken(subject, sealed)
	if err != nil {
		return "", false, err
	}
	return pt, true, nil
}

// OpenRoleConnectionToken opens a sealed role-connection token read
// elsewhere: account deletion reads the column inside its transaction,
// and the identity row is gone by the time the token is used.
func (s *SQLStore) OpenRoleConnectionToken(subject string, sealed []byte) (string, error) {
	if s.sealer == nil {
		return "", ErrNoSealer
	}
	pt, err := s.sealer.Open(sealed, roleConnectionAD(subject))
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// ClearRoleConnectionToken forgets the role-connection token: the person
// is unlinked until they link again. No identity is not an error.
func (s *SQLStore) ClearRoleConnectionToken(ctx context.Context, subject string) error {
	if _, err := s.db.ExecContext(ctx,
		`UPDATE identities SET role_connection_token = NULL WHERE provider = ? AND subject = ?`,
		ProviderDiscord, subject); err != nil {
		return fmt.Errorf("users: clear role-connection token: %w", err)
	}
	return nil
}
