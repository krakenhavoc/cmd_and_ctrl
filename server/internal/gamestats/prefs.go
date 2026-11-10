package gamestats

// The advanced-stats setting (ADR 0217 §2, owner answer 1; slice 3 of
// #4880). Basic stats are on for every account. Advanced stats are a
// free per-person switch, stored in migration 0030's stats_prefs, and
// a person who never touched it follows the server's default: the
// admin's stored default (stats_settings), or AdvancedDefault when no
// admin has set one.

import (
	"context"
	"errors"
	"time"
)

// AdvancedDefault is the fallback default when no admin has stored one
// (owner answer 1): off. Advanced stats are opt-in.
const AdvancedDefault = false

// SetBy is who last set a person's advanced-stats switch.
type SetBy string

const (
	// SetBySelf: the person switched it themselves.
	SetBySelf SetBy = "self"
	// SetByAdmin: an admin switched it for them (PUT
	// /admin/users/{id}/advanced-stats). The person can still switch
	// it off.
	SetByAdmin SetBy = "admin"
)

// Prefs is one person's stats setting. Advanced nil means "follow the
// default", and SetBy is then empty.
type Prefs struct {
	Advanced  *bool
	SetBy     SetBy
	UpdatedAt time.Time
	// ShareLinkAt is when the person's share link was made (ADR 0217
	// §6, slice 6), zero while they have none. The token itself is
	// never stored, only its hash.
	ShareLinkAt time.Time
	// SharePods is whether the person shares their stats with the
	// members of every pod they are in (ADR 0217 §6, ADR 0220 §7): one
	// switch for all pods, off until they turn it on.
	SharePods bool
}

// Effective is whether advanced stats are on for this person, given
// the server's default: their own answer when they or an admin gave
// one, else the default. So anyone who switched it off stays off when
// the default flips.
func (p Prefs) Effective(def bool) bool {
	if p.Advanced != nil {
		return *p.Advanced
	}
	return def
}

// DefaultSetting is the server's advanced-stats default. Stored is
// false when no admin has set one, and Advanced is then
// AdvancedDefault.
type DefaultSetting struct {
	Advanced  bool
	Stored    bool
	UpdatedAt time.Time
}

// AccountPrefs is one person's row in the admin listing: an account
// whose switch someone set. Every other account follows the default.
type AccountPrefs struct {
	UserID string
	Prefs
}

// ErrNoUser: the setting names an account that does not exist.
var ErrNoUser = errors.New("gamestats: no such user")

// ErrNoShareLink: no live account has a share link with that hash.
var ErrNoShareLink = errors.New("gamestats: no such share link")

// PrefsStore keeps the advanced-stats setting. Safe for concurrent use.
type PrefsStore interface {
	// Prefs is userID's setting; a person with no row follows the
	// default (a zero Prefs).
	Prefs(ctx context.Context, userID string) (Prefs, error)
	// SetAdvanced stores userID's switch and who set it. ErrNoUser if
	// the account does not exist.
	SetAdvanced(ctx context.Context, userID string, on bool, by SetBy, now time.Time) (Prefs, error)
	// SetSharePods stores userID's pod-sharing switch (ADR 0220 §7),
	// leaving the advanced switch and the share link alone. ErrNoUser
	// if the account does not exist.
	SetSharePods(ctx context.Context, userID string, on bool, now time.Time) (Prefs, error)
	// Default is the server's default.
	Default(ctx context.Context) (DefaultSetting, error)
	// SetDefault stores the server's default.
	SetDefault(ctx context.Context, on bool, now time.Time) (DefaultSetting, error)
	// Accounts lists every account whose switch is set, newest change
	// first, at most limit rows (<= 0: MaxPrefsAccounts).
	Accounts(ctx context.Context, limit int) ([]AccountPrefs, error)
	// SetShareLink stores the hash of userID's new share link (ADR 0217
	// §6), replacing any earlier one, which stops working. ErrNoUser if
	// the account does not exist. It leaves the advanced switch alone.
	SetShareLink(ctx context.Context, userID, hash string, now time.Time) (Prefs, error)
	// ClearShareLink revokes userID's share link; having none is not an
	// error.
	ClearShareLink(ctx context.Context, userID string, now time.Time) error
	// ShareLinkOwner is the live account whose share link has this
	// hash, or ErrNoShareLink.
	ShareLinkOwner(ctx context.Context, hash string) (string, error)
}

// MaxPrefsAccounts caps the admin listing, as adminview.MaxAccounts
// caps GET /admin/users.
const MaxPrefsAccounts = 1000
