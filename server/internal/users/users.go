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
	"fmt"
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
	// /c2-end host check: given the Discord user who ran the command,
	// which users(id) — if any — does games.created_by need to equal.
	UserIDForDiscord(ctx context.Context, discordID string) (uuid.UUID, error)
	// LastDeck reads the last deck the user seated (ADR 0110 section
	// 5, users.last_deck). The zero LastDeck means none has been
	// recorded. ErrNotFound if there is no such user.
	LastDeck(ctx context.Context, id uuid.UUID) (LastDeck, error)
	// SetLastDeck records the deck the user just seated. It refuses a
	// LastDeck that fails Validate (ErrInvalidLastDeck), and returns
	// ErrNotFound for an unknown user.
	SetLastDeck(ctx context.Context, id uuid.UUID, d LastDeck) error
	// Names reads the display name and avatar of each listed user in
	// one query, for the admin views' Live now (ADR 0124 §5). Only
	// ID, DisplayName and AvatarURL are set on the Users it returns.
	// An unknown id is absent from the map, not an error.
	Names(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]User, error)
}

// LastDeck kinds. A pre-built deck is not a decks row, so users.last_deck
// is JSON naming either, not a foreign key.
const (
	LastDeckLibrary  = "library"
	LastDeckPrebuilt = "prebuilt"
)

// MaxLastDeckIDLen bounds LastDeck.ID. A uuid is 36 bytes and a
// pre-built deck's slug is shorter.
const MaxLastDeckIDLen = 64

// ErrInvalidLastDeck is returned by SetLastDeck for a kind or id it
// will not store.
var ErrInvalidLastDeck = errors.New("users: invalid last deck")

// LastDeck is users.last_deck: {"kind": "library" | "prebuilt", "id": "..."}.
// The zero value means "none".
type LastDeck struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// IsZero reports whether no deck is recorded.
func (d LastDeck) IsZero() bool { return d == LastDeck{} }

// Validate checks a LastDeck is storable: a known kind, and an id that
// is non-empty and at most MaxLastDeckIDLen bytes. A library id must be
// a uuid.
func (d LastDeck) Validate() error {
	if d.Kind != LastDeckLibrary && d.Kind != LastDeckPrebuilt {
		return fmt.Errorf("%w: kind %q", ErrInvalidLastDeck, d.Kind)
	}
	if d.ID == "" || len(d.ID) > MaxLastDeckIDLen {
		return fmt.Errorf("%w: id must be 1 to %d bytes", ErrInvalidLastDeck, MaxLastDeckIDLen)
	}
	if d.Kind == LastDeckLibrary {
		if _, err := uuid.Parse(d.ID); err != nil {
			return fmt.Errorf("%w: a library id is a uuid", ErrInvalidLastDeck)
		}
	}
	return nil
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

// LastDeck reports ErrNotFound: there are no users without a store.
func (NoStore) LastDeck(context.Context, uuid.UUID) (LastDeck, error) {
	return LastDeck{}, ErrNotFound
}

// SetLastDeck always reports ErrNotFound, like Get.
func (NoStore) SetLastDeck(context.Context, uuid.UUID, LastDeck) error {
	return ErrNotFound
}

// Names implements Store: there are no users, so it knows no names.
func (NoStore) Names(context.Context, []uuid.UUID) (map[uuid.UUID]User, error) {
	return map[uuid.UUID]User{}, nil
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
