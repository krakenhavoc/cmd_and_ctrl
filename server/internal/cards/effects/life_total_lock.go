package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_total_lock.go — the card-facing half of CR 119.7 / CR 119.8
// (#1200, ADR 0085): a player whose LIFE TOTAL CAN'T CHANGE.
//
// Two shapes, and a card uses exactly one of them — the same split
// player_keywords.go describes for a player's abilities, because it is
// the same split and the same argument:
//
//   - a permanent's printed static, "Your life total can't change"
//     (Platinum Emperion) — declare Spec.PlayerLifeTotalLocked and
//     write no code at all. The engine reads the battlefield on every
//     query, so two of them compose and one leaving cannot revoke the
//     other's lock.
//   - a resolved spell, "until your next turn, your life total can't
//     change" (Teferi's Protection, Teferi's Reproach) — LockLifeTotal
//     below, which stores the lock with its CR 611.2 duration because
//     the spell is in exile a moment after it resolves.
//
// WHAT THE LOCK STOPS, once it is up: every life change in the game,
// because #482 routed every writer of a life total through one CR 614
// window — a GainLife, a drain, "each opponent loses 3 life", the
// lifelink credit, a life exchange, a Phyrexian symbol's two life, a
// shockland's two life. A COST that would pay life is refused outright
// (CR 119.8, CR 614.17b), at the validator as well as at the payment,
// so a bot is never offered the line and a client never shows a button
// that can only fail.
//
// WHAT IT DOES NOT STOP: the DAMAGE. Damage is still dealt — the
// event fires, "whenever ~ is dealt damage" triggers see it, the
// CR 903.10a commander tally accrues — and only the life loss CR 120.3a
// would have caused does not happen. Nor does it stop counters: poison
// still lands. Nor does it stop the player LOSING the game by any of
// CR 104.3's other routes.

// LockLifeTotal is "<player>'s life total can't change until
// <duration>" — the GRANTED half.
//
//	LockLifeTotal{
//	    Player:   ctx.Controller(),
//	    Label:    "Teferi's Protection — your life total can't change",
//	    Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
//	}.Apply(ctx)
//
// The Duration's zero value is "until end of turn", so a card file
// that forgets one gets the shortest window rather than a permanent
// lock — the same direction every other default in this catalog errs
// in.
//
// For a PERMANENT's printed "Your life total can't change", use
// Spec.PlayerLifeTotalLocked instead: that half is derived from the
// battlefield and needs no code.
type LockLifeTotal struct {
	// Player is the seat whose total is locked. Zero is a no-op.
	Player uuid.UUID
	// Label is the log's attribution, "<card> — <clause>".
	Label string
	// Duration is the CR 611.2 window. Build it with the constructors
	// in durations.go.
	Duration game.Duration
}

func (l LockLifeTotal) Apply(ctx *Context) error {
	if ctx == nil || l.Player == uuid.Nil {
		return nil
	}
	ctx.Game.GrantLifeTotalLockForEffect(l.Player, l.Label, ctx.Source(), l.Duration)
	return nil
}
