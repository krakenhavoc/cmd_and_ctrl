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
	// DiscordSubject is the Discord snowflake of a user's Discord
	// identity — identities.subject for provider = "discord" (ADR 0051
	// decision 2). It is what the DM-invite route needs to address a
	// person on Discord, and the only place our user id is turned back
	// into a snowflake. ErrNotFound when there is no such user, or
	// when the user has no Discord identity (which cannot happen while
	// Discord is the only provider, but will once there is a second).
	//
	// It is never part of a response body: a tablemate is offered by
	// OUR id, and the snowflake is resolved server-side.
	DiscordSubject(ctx context.Context, id uuid.UUID) (string, error)
	// UserIDForDiscord looks up the user linked to a Discord
	// snowflake (identities.provider = "discord", .subject =
	// discordID). ErrNotFound if no identity row matches — an unknown
	// snowflake, or one that never signed in. Added for #1098's
	// /cc-end host check: given the Discord user who ran the command,
	// which users(id) — if any — does games.created_by need to equal.
	UserIDForDiscord(ctx context.Context, discordID string) (uuid.UUID, error)
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

// DiscordSubject always reports ErrNotFound, for the same reason Get
// does: with no database there are no users to resolve.
func (NoStore) DiscordSubject(context.Context, uuid.UUID) (string, error) {
	return "", ErrNotFound
}

// UserIDForDiscord always reports ErrNotFound: there is nothing to
// look up without a store. Callers (the #1098 host check) treat that
// as "no", not as an error — see gameCreator's doc comment.
func (NoStore) UserIDForDiscord(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, ErrNotFound
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
