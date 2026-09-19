// Package users is the people half of the persistent store (ADR 0051
// decision 2, S34 sub-PR 2): a users row is a person, minted by us
// with our own uuid, and an identities row is one external account —
// a Discord login today — attached to it.
//
// The Discord snowflake is deliberately not the user's key. A second
// identity provider later is a new `provider` value on identities and
// nothing else; every seat, game and deck that points at a person
// points at users.id and never has to be re-keyed.
//
// Two Store implementations:
//
//   - SQLStore, over internal/db, whenever CMDCTRL_DATA_DIR is set.
//   - NoStore, when it is not. Sign-in still works and still mints a
//     session; it just carries a zero UserID, exactly as every session
//     did before this package existed.
package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// ProviderDiscord is identities.provider for a Discord login.
const ProviderDiscord = "discord"

// ErrNotFound is returned by Get for an unknown user.
var ErrNotFound = errors.New("users: not found")

// User is one users row.
type User struct {
	// ID is the user's uuid. uuid.Nil only from NoStore, meaning "no
	// user store on this deployment".
	ID          uuid.UUID
	DisplayName string
	// AvatarURL is the same-origin avatar path the client renders
	// (/avatars/<snowflake>/<hash>.png), or "" for none.
	AvatarURL  string
	CreatedAt  time.Time
	LastSeenAt time.Time
	// SessionsInvalidBefore is decision 6's revocation watermark: a
	// session for this user issued at or before it is refused. Zero
	// until the user first signs out everywhere or is removed by an
	// admin. The request path reads it through Revocations, never
	// through Get; see there.
	SessionsInvalidBefore time.Time
}

// Store reads and writes people. Implementations are safe for
// concurrent use.
type Store interface {
	// UpsertFromDiscord records a Discord sign-in and returns the user
	// it belongs to. The first sign-in for a snowflake mints a new
	// user and links the identity; a later one refreshes the display
	// name, avatar and last_seen_at on both rows.
	//
	// refreshToken is Discord's OAuth refresh token. It is stored
	// only encrypted (see Sealer); a store with no key discards it
	// and records NULL. An empty refreshToken leaves any stored one
	// in place. scopes is the space-separated scope string Discord
	// granted.
	UpsertFromDiscord(ctx context.Context, profile discord.User, refreshToken, scopes string) (User, error)
	// Get reads one user. ErrNotFound if there is none.
	Get(ctx context.Context, id uuid.UUID) (User, error)
}

// NoStore is the Store for a deployment with no database. Every
// sign-in returns the zero User, so the principal it mints carries a
// zero UserID — the pre-S34 behaviour, unchanged.
type NoStore struct{}

// UpsertFromDiscord records nothing and returns the zero User.
func (NoStore) UpsertFromDiscord(context.Context, discord.User, string, string) (User, error) {
	return User{}, nil
}

// Get always reports ErrNotFound: there are no users without a store.
func (NoStore) Get(context.Context, uuid.UUID) (User, error) {
	return User{}, ErrNotFound
}

// AvatarPath is the same-origin path the client loads a Discord
// avatar from (served by the avatar cache at /avatars/{id}/{hash}),
// or "" when the account has no custom avatar. It matches what the
// client builds from a seat's Discord fields, so either source
// renders the same image.
func AvatarPath(discordID, avatarHash string) string {
	if discordID == "" || avatarHash == "" {
		return ""
	}
	return "/avatars/" + discordID + "/" + avatarHash + ".png"
}
