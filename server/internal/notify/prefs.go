// Package notify is ADR 0162's Discord DM notifications: a signed-in
// player who opts in is sent one direct message when a game has waited
// on them for two minutes while they are away, or when a table they sit
// at fills or starts while they are away.
//
// Two halves:
//
//   - the preferences (this file, store_sql.go): three switches per
//     person, every one off by default, and when Discord last refused a
//     DM to them. Their own table (migration 0022), because the
//     user_settings body is opaque to the server (ADR 0110 §4).
//   - the Notifier (notifier.go): a goroutine that looks at the live
//     tables every few seconds, keeps the per-seat clocks and the
//     per-player rate window in memory, and sends through the server's
//     bot token (discord.Bot.SendDM). It is fed by the lobby (a table
//     filled or started), the ws hub (who is away) and the rooms (who
//     the engine waits on, who last acted).
//
// With no bot token the whole feature is off: main builds no Notifier,
// and the routes answer 503.
package notify

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// Kind is one of the three switches, and the kind label of
// cmdctrl_dm_notifications_total.
type Kind string

const (
	// KindTurn: the game has waited on you for TurnWait while you are
	// away.
	KindTurn Kind = metrics.DMKindTurn
	// KindTableReady: a table you sit at has filled or started while
	// you are away.
	KindTableReady Kind = metrics.DMKindTableReady
	// KindTableFilled: a table you host has filled while you are away.
	KindTableFilled Kind = metrics.DMKindTableFilled
)

// ErrNoStore is what NoStore's writes answer.
var ErrNoStore = errors.New("notify: no store configured")

// Prefs is one person's notification_prefs row. The zero value is the
// default: every switch off, nothing refused.
type Prefs struct {
	Turn        bool
	TableReady  bool
	TableFilled bool
	// DMFailedAt is when Discord last answered 50007 for this person,
	// zero for never. While it is set nothing is sent.
	DMFailedAt time.Time
}

// Wants reports whether the switch for k is on and Discord has not
// refused this person.
func (p Prefs) Wants(k Kind) bool {
	if !p.DMFailedAt.IsZero() {
		return false
	}
	switch k {
	case KindTurn:
		return p.Turn
	case KindTableReady:
		return p.TableReady
	case KindTableFilled:
		return p.TableFilled
	}
	return false
}

// Update is a PUT /me/notifications body: nil leaves a switch as it is.
type Update struct {
	Turn        *bool
	TableReady  *bool
	TableFilled *bool
}

// Empty reports whether u changes nothing.
func (u Update) Empty() bool { return u.Turn == nil && u.TableReady == nil && u.TableFilled == nil }

// Store reads and writes the preferences. Safe for concurrent use.
type Store interface {
	// Get reads one person's switches. No row is the zero Prefs and no
	// error.
	Get(ctx context.Context, user uuid.UUID) (Prefs, error)
	// Put applies u and returns the stored row. Turning any switch on
	// that was off clears DMFailedAt (ADR 0162 §4: "turning any switch
	// off and on again clears it").
	Put(ctx context.Context, user uuid.UUID, u Update) (Prefs, error)
	// Active lists everyone with at least one switch on and no refusal
	// on record: the only people the Notifier may message.
	Active(ctx context.Context) (map[uuid.UUID]Prefs, error)
	// MarkFailed records Discord's 50007 for user at at.
	MarkFailed(ctx context.Context, user uuid.UUID, at time.Time) error
}

// NoStore is the Store for a deployment with no database: nobody has a
// switch on, and nothing can be saved.
type NoStore struct{}

// Get always answers the defaults.
func (NoStore) Get(context.Context, uuid.UUID) (Prefs, error) { return Prefs{}, nil }

// Put always fails.
func (NoStore) Put(context.Context, uuid.UUID, Update) (Prefs, error) { return Prefs{}, ErrNoStore }

// Active is always empty.
func (NoStore) Active(context.Context) (map[uuid.UUID]Prefs, error) { return nil, nil }

// MarkFailed has nowhere to write.
func (NoStore) MarkFailed(context.Context, uuid.UUID, time.Time) error { return ErrNoStore }

var _ Store = NoStore{}
